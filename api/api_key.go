package api

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

// 假设你有一个全局的 mutex 来保护对 token 文件的读写，以防止并发问题
var tokenFileMutex sync.Mutex

// keyStatusKey 是存储 key 状态的 map 的键类型
type keyStatusKey string

// keyStatusValue 是存储 key 状态的 map 的值类型
type keyStatusValue bool

// keyStatusMap 用于存储 key 的使用状态
var keyStatusMap = make(map[keyStatusKey]keyStatusValue)

// keyStatusMutex 用于保护 keyStatusMap 的并发访问
var keyStatusMutex sync.Mutex

// keyStatusCond 是条件变量，用于在 key 状态变化时通知等待者
var keyStatusCond = sync.NewCond(&keyStatusMutex)

// 为了演示，我们将随机数种子初始化放在 init 函数中
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetAllKeysFromFile 读取文件中的所有有效秘钥
// 注意: 这个函数假设 key 文件在程序运行时不会变化。
// 如果文件会变化，需要更复杂的机制来重新加载和同步 key 列表。
func getAllKeysFromFile(keyFilePath string) ([]string, error) {
	file, err := os.Open(keyFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("key file not found at: %s", keyFilePath)
		}
		return nil, fmt.Errorf("failed to open key file: %w", err)
	}
	defer file.Close()

	var keys []string
	reader := bufio.NewReader(file)
	for {
		// ReadLine 读取一行，忽略换行符
		// 注意: ReadLine 可能会返回切片，直到下一次调用前都有效。
		// 对于长期存储或修改，应将 bytes 复制为 string。
		// 这里因为是逐行处理，且只用于比较和添加到切片，直接使用 string 转换是安全的。
		lineBytes, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read key file line: %w", err)
		}
		key := strings.TrimSpace(string(lineBytes))
		if key != "" {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// GetRandomKey 从指定的 key 文件中随机读取一个未被锁定的秘钥。
// 如果当前没有可用秘钥，则会等待直到有秘钥被释放。
// keyFilePath: key 文件的路径
// 返回值:
// selectedKey: 获取到的秘钥
// err: 如果发生文件读取错误等，则返回错误
func GetRandomKey(keyFilePath string, falseKeys []string) (selectedKey string, err error) {
	// 在访问 statusMap 之前锁定 mutex
	keyStatusMutex.Lock()
	defer keyStatusMutex.Unlock() // 在函数退出时解锁

	// 读取所有 key，以便知道总共有多少 key
	// 注意：如果 key 文件内容可能变化，这里需要更复杂的处理
	allKeys, err := getAllKeysFromFile(keyFilePath)
	if err != nil {
		// 即使文件读取失败，也需要解锁，否则会导致死锁
		fmt.Errorf("failed to read all keys from file: %w", err)
	}

	if len(allKeys) == 0 {
		// 如果文件中根本没有 key，直接返回错误，避免无限等待
		fmt.Errorf("no valid keys found in the file: %s", keyFilePath)
	}

	// 循环直到找到一个可用的 key
	for {
		var availableKeys []string // 存储未被使用的秘钥

		// 遍历所有 key，查找未被使用的
		for _, key := range allKeys {
			// 检查 map 中是否存在该 key，并且它的 value 是 false (未被使用)
			if !keyStatusMap[keyStatusKey(key)] {
				availableKeys = append(availableKeys, key) // 如果未被使用，添加到可用列表
			}
		}

		if len(availableKeys) > 0 {
			// 如果有可用 key，随机选择一个
			randomIndex := rand.Intn(len(availableKeys))
			selectedKey = availableKeys[randomIndex]

			// 将选中的 key 标记为正在使用
			keyStatusMap[keyStatusKey(selectedKey)] = true

			// 返回获取到的 key
			return selectedKey, nil
		}

		// 如果没有可用 key，释放锁并等待条件变量
		// Wait 方法会自动释放锁并在收到 Signal 或 Broadcast 时再次获取锁
		fmt.Printf("所有Key都在使用中,请耐心等待其他 Key 值释放 (Total keys: %d)\n", len(allKeys))

		//// 将falseKeys的值拿一个出来
		//if len(falseKeys) > 0 {
		//	firstKey := falseKeys[0] // 取出第一个元素
		//	fmt.Println("从列表中取出的一个key值强行解锁：", firstKey)
		//	ReleaseKey(firstKey)
		//} else {
		//	fmt.Println("falseKeys 列表为空，无法取出值。")
		//}

		keyStatusCond.Wait() // 在这里等待，直到有key被释放并广播
		// 当 Wait 返回时，表示有一个 key 状态发生变化，并且我们再次获取了锁。
		// 循环会继续，重新检查是否有可用 key。
		// 注意：Wait 返回后，你需要重新检查条件，因为可能有多个goroutine被唤醒，
		// 并且在你的goroutine重新获取锁之前，其他goroutine已经获取了key。
	}
}

// ReleaseKey 释放一个正在使用的秘钥，并通知等待者
// key: 需要释放的秘钥字符串
func ReleaseKey(key string) {
	// 在访问 statusMap 之前锁定 mutex
	keyStatusMutex.Lock()
	defer keyStatusMutex.Unlock() // 在函数退出时解锁

	// 将指定的 key 标记为未在使用
	// 检查 key 是否真的被标记为使用中，更严谨
	if keyStatusMap[keyStatusKey(key)] {
		delete(keyStatusMap, keyStatusKey(key)) // 使用 delete 从 map 中移除 key
		// 释放 key 后，通知一个或所有等待 GetRandomKey 的 goroutine
		// keyStatusCond.Signal() // 通知一个等待者
		keyStatusCond.Broadcast() // 通知所有等待者，如果你希望所有等待者都尝试获取key
		fmt.Printf("Released key: %s. Notifying waiters.\n", key)
	} else {
		fmt.Printf("Attempted to release a key that was not in use: %s\n", key)
	}
}

// GetLockedKeys 用于获取当前被锁定的key列表 (可选，仅用于调试或监控)
func GetLockedKeys() []string {
	keyStatusMutex.Lock()
	defer keyStatusMutex.Unlock()

	var lockedKeys []string
	for key, inUse := range keyStatusMap {
		if inUse {
			lockedKeys = append(lockedKeys, string(key))
		}
	}
	return lockedKeys
}

// HandleUnauthorizedKey 将给定的 token 从 tokenFile 中删除，并添加到 errorTokenFile 中
func HandleUnauthorizedKey(token string) error {
	tokenFileMutex.Lock()
	defer tokenFileMutex.Unlock()

	tokenFile := viper.GetString("Nkey.path")
	errorTokenFile := viper.GetString("Nkey.path_err")

	// 从 tokens 文件中读取所有 token
	tokens, err := readTokens(tokenFile)
	if err != nil {
		return fmt.Errorf("failed to read tokens from %s: %w", tokenFile, err)
	}

	// 找到并删除未授权的 token
	foundAndRemoved := false
	var updatedTokens []string
	for _, t := range tokens {
		if t == token {
			foundAndRemoved = true
			log.Printf("Removing unauthorized key from %s: %s", tokenFile, token)
		} else {
			updatedTokens = append(updatedTokens, t)
		}
	}

	if !foundAndRemoved {
		log.Printf("Warning: Unauthorized key not found in %s: %s", tokenFile, token)
		// 如果未找到，可能已经处理过了，或者 Key 来源有问题
		// 这里可以选择是返回错误还是继续添加到 error 文件，取决于你的需求
	}

	// 将更新后的 token 写回 tokens 文件
	err = writeTokens(tokenFile, updatedTokens)
	if err != nil {
		return fmt.Errorf("failed to write updated tokens to %s: %w", tokenFile, err)
	}

	// 将未授权的 token 追加到 errorTokens 文件
	err = appendTokenToFile(errorTokenFile, token)
	if err != nil {
		return fmt.Errorf("failed to append unauthorized token to %s: %w", errorTokenFile, err)
	}

	return nil
}

// readTokens 从指定文件读取所有非空行作为 token
func readTokens(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		// 文件不存在，返回空列表，这不是错误
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	var tokens []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		token := strings.TrimSpace(scanner.Text())
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}
	return tokens, nil
}

// writeTokens 将给定的 token 列表写入到指定文件，覆盖现有内容
func writeTokens(filename string, tokens []string) error {
	content := strings.Join(tokens, "\n") + "\n"             // 每行一个 token
	err := ioutil.WriteFile(filename, []byte(content), 0644) // 0644 是标准的文件权限
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	return nil
}

// appendTokenToFile 将给定的 token 追加到指定文件的末尾
func appendTokenToFile(filename string, token string) error {
	// O_APPEND 以追加模式打开，O_CREATE 如果文件不存在则创建，O_WRONLY 只写入
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s for appending: %w", filename, err)
	}
	defer file.Close()

	_, err = file.WriteString(token + "\n") // 追加 token 并换行
	if err != nil {
		return fmt.Errorf("failed to write token to file %s: %w", filename, err)
	}
	return nil
}
