package chezmoi

var _ PersistentState = newTestPersistentState()

type testPersistentState map[string]map[string][]byte

func newTestPersistentState() testPersistentState {
	return make(testPersistentState)
}

func (s testPersistentState) Delete(bucket, key []byte) error {
	bucketMap, ok := s[string(bucket)]
	if !ok {
		return nil
	}
	delete(bucketMap, string(key))
	return nil
}

func (s testPersistentState) Get(bucket, key []byte) ([]byte, error) {
	bucketMap, ok := s[string(bucket)]
	if !ok {
		return nil, nil
	}
	return bucketMap[string(key)], nil
}

func (s testPersistentState) Set(bucket, key, value []byte) error {
	bucketMap, ok := s[string(bucket)]
	if !ok {
		bucketMap = make(map[string][]byte)
		s[string(bucket)] = bucketMap
	}
	bucketMap[string(key)] = value
	return nil
}
