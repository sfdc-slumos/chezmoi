package chezmoi

// A PersistentState is a persistent state.
type PersistentState interface {
	Get(bucket, key []byte) ([]byte, error)
	Delete(bucket, key []byte) error
	Set(bucket, key, value []byte) error
}

type nullPersistentState struct{}

func (nullPersistentState) Delete(bucket, key []byte) error        { return nil }
func (nullPersistentState) Get(bucket, key []byte) ([]byte, error) { return nil, nil }
func (nullPersistentState) Set(bucket, key, value []byte) error    { return nil }
