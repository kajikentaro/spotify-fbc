package service_compares

type DiffState int

const (
	LocalOnly DiffState = iota + 1
	RemoteOnly
	Both
)

type WithDiffState[T any] struct {
	V         T
	DiffState DiffState
}

func calcDiff[T any](local []T, remote []T, getId func(T) string, merge func(local T, remote T) T) []WithDiffState[T] {
	res := []WithDiffState[T]{}

	idToPlaylist := map[string]WithDiffState[T]{}
	for _, v := range remote {
		id := getId(v)
		idToPlaylist[id] = WithDiffState[T]{V: v, DiffState: RemoteOnly} // まずRemoteOnlyで登録しておく
	}
	for _, v := range local {
		id := getId(v)
		if id == "" {
			// ユーザーが作成してSyncされていないプレイリストはIDが空になる
			// この時点でLocalOnlyが確定するので、resに追加
			res = append(res, WithDiffState[T]{V: v, DiffState: LocalOnly})
			continue
		}

		remote, ok := idToPlaylist[id]
		if ok {
			// Remoteにも存在する場合
			idToPlaylist[id] = WithDiffState[T]{V: merge(v, remote.V), DiffState: Both}
		} else {
			// Remoteに存在しない場合
			idToPlaylist[id] = WithDiffState[T]{V: v, DiffState: LocalOnly}
		}
	}

	// LocalOnly -> RemoteOnly -> Bothの順でresに追加する
	for _, v := range idToPlaylist {
		if v.DiffState == LocalOnly {
			res = append(res, v)
		}
	}
	for _, v := range idToPlaylist {
		if v.DiffState == RemoteOnly {
			res = append(res, v)
		}
	}
	for _, v := range idToPlaylist {
		if v.DiffState == Both {
			res = append(res, v)
		}
	}
	return res
}
