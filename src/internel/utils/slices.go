package utils

import (
	"math/rand"
	"time"
)

func RemoveRepeatedElement[S ~[]E, E comparable](s S) S {
	result := make([]E, 0)
	m := make(map[E]bool)
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}

// ShuffledCopy 泛型函数，用于复制并打乱切片
// allowDuplicates 参数控制是否允许重复
// 如果允许重复，将随机选择-2, -1, 0, 1, 2作为长度变化量
func ShuffledCopy[T any](slice []T, allowDuplicates bool) []T {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 初始化新切片长度为原切片长度
	newLen := len(slice)

	if allowDuplicates {
		// 从 -2, -1, 0, 1, 2 中随机选择一个作为长度变化量
		deltaOptions := []int{-2, -1, 0, 1, 2}
		delta := deltaOptions[rand.Intn(len(deltaOptions))]

		// 计算新的切片长度，确保不小于 0
		newLen += delta
		if newLen < 0 {
			newLen = 0
		}
	}

	// 创建新切片
	newSlice := make([]T, newLen)

	if allowDuplicates {
		// 允许重复时，从原切片中随机选择元素填充新切片
		for i := 0; i < newLen; i++ {
			newSlice[i] = slice[rand.Intn(len(slice))]
		}
		return newSlice
	}
	// 不允许重复时，直接复制原切片并打乱顺序
	copy(newSlice, slice)

	// 打乱新切片
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(
		len(newSlice), func(i, j int) {
			newSlice[i], newSlice[j] = newSlice[j], newSlice[i]
		},
	)
	return newSlice
}
