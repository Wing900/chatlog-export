package darwin

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"os"

	"github.com/sjzar/chatlog/internal/errors"
	"github.com/sjzar/chatlog/internal/wechat/decrypt/common"

	"golang.org/x/crypto/pbkdf2"
)

// 常量定义
const (
	V3PageSize   = 1024
	HmacSHA1Size = 20
)

// V3Decryptor 实现 macOS V3 版本的解密器
type V3Decryptor struct {
	// macOS V3 特定参数
	hmacSize int
	hashFunc func() hash.Hash
	reserve  int
	pageSize int
	version  string
}

// NewV3Decryptor 创建 macOS V3 解密器
func NewV3Decryptor() *V3Decryptor {
	hashFunc := sha1.New
	hmacSize := HmacSHA1Size
	reserve := common.IVSize + hmacSize
	if reserve%common.AESBlockSize != 0 {
		reserve = ((reserve / common.AESBlockSize) + 1) * common.AESBlockSize
	}

	return &V3Decryptor{
		hmacSize: hmacSize,
		hashFunc: hashFunc,
		reserve:  reserve,
		pageSize: V3PageSize,
		version:  "macOS v3",
	}
}

// deriveKeys 派生 MAC 密钥
// 注意：macOS V3 版本直接使用提供的密钥作为加密密钥，不进行 PBKDF2 派生
func (d *V3Decryptor) deriveKeys(key []byte, salt []byte) ([]byte, []byte) {
	// 对于 macOS V3，直接使用密钥作为加密密钥
	encKey := key

	// 生成 MAC 密钥
	macSalt := common.XorBytes(salt, 0x3a)
	macKey := pbkdf2.Key(encKey, macSalt, 2, common.KeySize, d.hashFunc)

	return encKey, macKey
}

// Validate 验证密钥是否有效
func (d *V3Decryptor) Validate(page1 []byte, key []byte) bool {
	if len(page1) < d.pageSize || len(key) != common.KeySize {
		return false
	}

	salt := page1[:common.SaltSize]
	return common.ValidateKey(page1, key, salt, d.hashFunc, d.hmacSize, d.reserve, d.pageSize, d.deriveKeys)
}

// Decrypt 解密数据库
func (d *V3Decryptor) Decrypt(ctx context.Context, dbfile string, hexKey string, output io.Writer) error {
	// 解码密钥
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return errors.DecodeKeyFailed(err)
	}

	// 打开数据库文件并读取基本信息
	dbInfo, err := common.OpenDBFile(dbfile, d.pageSize)
	if err != nil {
		return err
	}

	// 验证密钥
	if !d.Validate(dbInfo.FirstPage, key) {
		return errors.ErrDecryptIncorrectKey
	}

	// 计算密钥
	encKey, macKey := d.deriveKeys(key, dbInfo.Salt)

	// 打开数据库文件
	dbFile, err := os.Open(dbfile)
	if err != nil {
		return errors.OpenFileFailed(dbfile, err)
	}
	defer dbFile.Close()

	// 写入 SQLite 头
	_, err = output.Write([]byte(common.SQLiteHeader))
	if err != nil {
		return errors.WriteOutputFailed(err)
	}

	// 处理每一页
	pageBuf := make([]byte, d.pageSize)

	for curPage := int64(0); curPage < dbInfo.TotalPages; curPage++ {
		// 检查是否取消
		select {
		case <-ctx.Done():
			return errors.ErrDecryptOperationCanceled
		default:
			// 继续处理
		}

		// 读取一页
		n, err := io.ReadFull(dbFile, pageBuf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// 处理最后一部分页面
				if n > 0 {
					break
				}
			}
			return errors.ReadFileFailed(dbfile, err)
		}

		// 检查页面是否全为零
		allZeros := true
		for _, b := range pageBuf {
			if b != 0 {
				allZeros = false
				break
			}
		}

		if allZeros {
			// 写入零页面
			_, err = output.Write(pageBuf)
			if err != nil {
				return errors.WriteOutputFailed(err)
			}
			continue
		}

		// 解密页面
		decryptedData, err := common.DecryptPage(pageBuf, encKey, macKey, curPage, d.hashFunc, d.hmacSize, d.reserve, d.pageSize)
		if err != nil {
			return err
		}

		// 写入解密后的页面
		_, err = output.Write(decryptedData)
		if err != nil {
			return errors.WriteOutputFailed(err)
		}
	}

	return nil
}

// GetPageSize 返回页面大小
func (d *V3Decryptor) GetPageSize() int {
	return d.pageSize
}

// GetReserve 返回保留字节数
func (d *V3Decryptor) GetReserve() int {
	return d.reserve
}

// GetHMACSize 返回HMAC大小
func (d *V3Decryptor) GetHMACSize() int {
	return d.hmacSize
}

// GetVersion 返回解密器版本
func (d *V3Decryptor) GetVersion() string {
	return d.version
}
