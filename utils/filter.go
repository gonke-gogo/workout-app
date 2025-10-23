package utils

// Filter ジェネリクスを使用したフィルタリング関数
// 任意の型のスライスから条件を満たす要素だけを抽出する汎用的な処理
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}
