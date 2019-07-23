package main // Might need to change package to proj2 to pass auto grader

//package proj2

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/ryanleh/cs161-p2/userlib"
	_ "strconv"
	"strings"
)

func someUsefulThings() {
	// Creates a random UUID
	userlib.DebugPrint = true
	f := uuid.New()
	userlib.DebugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	userlib.DebugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	userlib.DebugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	userlib.DebugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	userlib.DebugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	userlib.DebugMsg("Creation of error %v", errors.New(strings.ToTitle("This is an error")))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var pk userlib.PKEEncKey
	var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("Key is %v, %v", pk, sk)
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

// The structure definition for storing things on the Datastore.
type wrapper struct {
	cyphers [][]byte
	hmacs   [][]byte
}

/**
This is the main wrapper function that is used to ensure integrity of a slice of
cypher text (C0 .. Cn) when it is stored on the Datastore.

TODO: verify that pointer works for the slices since wrapper will get stored...

It takes:
	- A HMAC key byte slice.
	- A slice of byte slice cypher texts s.t. the first element is the IV byte slice.
It returns:
	- A 'wrapper' struct following the format described in the design doc.
	- A nil error if successful.
*/
func wrap(key *[]byte, cyphers *[][]byte) (wrap *wrapper, err error) {
	return
}

/**
This is the main unwrapping function to read encrypted cypher text from the Datastore
and check for integrity.

TODO: verify that pointer works for the slices since wrapper will get stored...

It takes:
	- A HMAC key byte slice.
	- A 'wrapper' struct following the format described in the design doc.
It returns:
	- A slice of byte slice cypher texts s.t. the first element is the IV byte slice.
	- A nil error if successful.
*/
func unwrap(key *[]byte, wrap *wrapper) (cyphers *[][]byte, err error) {
	return
}

/**
This function symmetrically encrypts a byte slice while handling the padding.

TODO: verify that pointer works for the slices since wrapper will get stored...

It takes:
	- A encryption key byte slice.
	- A IV byte slice. (size has to be equal to userlib.AESBlockSize)
	- A byte slice to be encrypted.
It returns:
	- A slice of byte slice cypher texts s.t. the first element is the IV byte slice.
	- A nil error if successful.
*/
func symEncrypt(key *[]byte, iv *[]byte, data *[]byte) (cyphers *[][]byte, err error) {
	return
}

/**
This function symmetrically decrypts slice of byte slice cypher texts and removes the padding.

TODO: verify that pointer works for the slices since wrapper will get stored...

It takes:
	- A encryption key byte slice.
	- A slice of byte slice cypher texts s.t. the first element is the IV byte slice.
It returns:
	- A byte slice of the unencrypted data
	- A nil error if successful.
*/
func symDecrypt(key *[]byte, cyphers *[][]byte) (data *[]byte, err error) {
	return
}

// The structure definition for a user record
type User struct {
	Username string

	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored
// User data: the name used in the datastore should not be guessable
// without also knowing the password and username.

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the user has a STRONG password
func InitUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata

	return &userdata, nil
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata

	return userdataptr, nil
}

// This stores a file in the datastore.
//
// The name of the file should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {
	return
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.

func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	return
}

// You may want to define what you actually want to pass as a
// sharingRecord to serialized/deserialize in the data store.
type sharingRecord struct {
}

// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.

func (userdata *User) ShareFile(filename string, recipient string) (
	magic_string string, err error) {

	return
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {
	return nil
}

// Removes access for all others.
func (userdata *User) RevokeFile(filename string) (err error) {
	return
}

// Example of how to use CBC encrypted that they give to return a list of cyphers and how to
// use C_i-1 and C_i to get P_i
func cbc_enc_ex() {
	// This example is given TestInit test.
	IV := userlib.RandomBytes(userlib.AESBlockSize)
	key := userlib.RandomBytes(userlib.AESBlockSize)
	msg := userlib.RandomBytes(3 * userlib.AESBlockSize)
	userlib.DebugPrint = true
	userlib.DebugMsg("IV: %x", IV)

	msg_list := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		msg_list[i] = msg[i*userlib.AESBlockSize : (i+1)*userlib.AESBlockSize]
	}
	userlib.DebugMsg("Msg: %x", msg)
	userlib.DebugMsg("Msg Blocks: %x", msg_list)

	enc := userlib.SymEnc(key, IV, msg)
	enc_list := make([][]byte, 4)
	for i := 0; i < 4; i++ {
		enc_list[i] = enc[i*userlib.AESBlockSize : (i+1)*userlib.AESBlockSize]
	}
	userlib.DebugMsg("ENC Msg: %x", enc)
	userlib.DebugMsg("ENC Msg Blocks: %x", enc_list)

	dec := userlib.SymDec(key, append(IV, enc...))[userlib.AESBlockSize:] // Note that the first block is garbage.
	userlib.DebugMsg("DEC Msg: %x", dec)

	dec_list := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		e_msg := append(enc_list[i], enc_list[i+1]...)
		//userlib.DebugMsg("e_msg %x", e_msg)
		d_msg := userlib.SymDec(key, e_msg)
		//userlib.DebugMsg("d_msg %x", d_msg)
		dec_list[i] = d_msg
	}

	userlib.DebugMsg("DEC Msg Blocks: %x", dec_list)
}

// Each test function can be stepped through here before moving it over to the test file.
func main() {
	someUsefulThings()
}
