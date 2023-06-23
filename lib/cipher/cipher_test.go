package cipher

type testUser struct {
	email string
	id    string
	nonce []byte
}

func (user *testUser) Mask() []byte {
	return []byte(user.id)
}

func (user *testUser) MasterKey() []byte {
	keyLen := len(user.email)
	if keyLen > MasterKeyLen {
		return []byte(user.email)[:MasterKeyLen]
	}

	if keyLen < MasterKeyLen {
		offset := MasterKeyLen - keyLen
		for i := 0; i < offset; i++ {
			user.email += " "
		}
	}
	return []byte(user.email)
}

func (user *testUser) Serialize(data, nonce []byte) {
	user.id = string(data)
	user.nonce = nonce
}

func (user *testUser) Deserialize() ([]byte, []byte) {
	return []byte(user.id), user.nonce
}

func testInit() *testUser {
	return &testUser{"foo@bar.com", "foobarbaz", nil}
}

func testInitWithData() (*testUser, *KDF) {
	kdf, _ := NewKDF()
	user := testInit()
	return user, kdf
}
