package proj2

import (
	_ "encoding/hex"
	"encoding/json"
	_ "errors"
	_ "github.com/google/uuid"
	"github.com/ryanleh/cs161-p2/userlib"
	"reflect"
	_ "strconv"
	"strings"
	"testing"
	"time"
)

/**
Real tests that should work for all implementations.
*/

// This assumes that each unique username will only call init once.
func TestInitAndGetBasics(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	bob, err := InitUser("bob", "fubar")
	if bob == nil || err != nil {
		t.Error(err)
		return
	}
	getBob, err := GetUser("bob", "fubar")
	if getBob == nil || err != nil {
		t.Error(err)
		return
	}

	bobBytes, _ := json.Marshal(bob)
	getBobBytes, _ := json.Marshal(getBob)
	if !reflect.DeepEqual(bobBytes, getBobBytes) {
		t.Error("Init and Get userdata are not the same.")
		return
	}

	_, err = GetUser("bob", "wrong")
	if err == nil {
		t.Error("Got a user that is suppose to not exist.")
		return
	}

	_, err = GetUser("wrong", "fubar")
	if err == nil {
		t.Error("Got a user that is suppose to not exist.")
		return
	}

	var keys []userlib.UUID
	var vals [][]byte
	for k, v := range datastore {
		keys = append(keys, k)
		vals = append(vals, v)
	}

	for val := range vals {
		if strings.Contains("bob", string(val)) || strings.Contains("alice", string(val)) {
			t.Error("Username is not obscured.")
			return
		}
	}

}

// This assumes that each unique username will only call init once.
func TestInitAndGetWithCorruptDatastore(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	_, err := InitUser("bob", "fubar")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = InitUser("alice", "fubar")
	if err != nil {
		t.Error(err)
		return
	}

	var keys []userlib.UUID
	var vals [][]byte
	for k, v := range datastore {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	userlib.DatastoreSet(keys[0], vals[1])
	for i := 1; i < len(keys); i++ {
		userlib.DatastoreSet(keys[i], vals[0])
	}

	_, err = GetUser("alice", "fubar")
	if err == nil {
		t.Error("Datastore was corrupted for alice but still got user.")
		return
	}
	_, err = GetUser("bob", "fubar")
	if err == nil {
		t.Error("Datastore was corrupted for bob but still got user.")
		return
	}

	userlib.DatastoreClear()
	userlib.KeystoreClear()

	_, err = InitUser("bob", "fubar")
	if err != nil {
		t.Error(err)
		return
	}
	userlib.DatastoreClear()
	_, err = GetUser("bob", "fubar")
	if err == nil {
		t.Error("Datastore was empty but still got user.")
		return
	}

	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore = userlib.DatastoreGetMap()

	_, err = InitUser("bob", "fubar")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = InitUser("alice", "fubar")
	if err != nil {
		t.Error(err)
		return
	}

	var keys1 []userlib.UUID
	var vals1 [][]byte
	for k, v := range datastore {
		keys1 = append(keys1, k)
		vals1 = append(vals1, v)
	}
	datastore[keys1[0]] = userlib.RandomBytes(len(keys1[0]))

	_, err0 := GetUser("bob", "fubar")
	_, err1 := GetUser("alice", "fubar")
	if err0 == nil && err1 == nil {
		t.Error("successfully got all users when datastore was corrupted.")
	}
}

func TestStorageBasic(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	fileNames := []string{"f1", "f2", "f3", "f4", "f5"}
	userNames := []string{"u1", "u2", "u3", "u4", "u5"}
	for i, offset := range []int{-4, -1, 0, 1, 7} {
		user, err1 := InitUser(userNames[i], "fubar")
		if err1 != nil {
			t.Error(err1)
			return
		}

		file := userlib.RandomBytes(userlib.AESBlockSize - offset)
		user.StoreFile(fileNames[i], file)

		// Get user to check for userdata update.
		user, err2 := GetUser(userNames[i], "fubar")
		if err2 != nil {
			t.Error(err2)
			return
		}

		loadedFile, err3 := user.LoadFile(fileNames[i])
		if err3 != nil {
			t.Error("Failed to upload and download", err3)
			return
		}
		if !reflect.DeepEqual(file, loadedFile) {
			t.Error("Loaded file is not the same original\n",
				file, loadedFile)
			return
		}

		someFile, err4 := user.LoadFile("bad")
		if err4 != nil {
			t.Error("Raised error on a load of a file that DNE.")
		}

		if someFile != nil {
			t.Error("Load of a file that DNE did not return nil.")
		}
	}

	// Test the file overwrite case in our implementation.
	file := userlib.RandomBytes(userlib.AESBlockSize - 7)
	user, err2 := GetUser(userNames[4], "fubar")
	user.StoreFile(fileNames[4], file)
	if err2 != nil {
		t.Error(err2)
		return
	}
	loadedFile, err3 := user.LoadFile(fileNames[4])
	if err3 != nil {
		t.Error("Failed to upload and download", err3)
		return
	}
	if !reflect.DeepEqual(file, loadedFile) {
		t.Log("StoreFile overwrite failed. This is acceptable.")
		// Some implementations don't implement overwrite so this is not a fail.
	}

	file = make([]byte, userlib.AESBlockSize)
	file[0] = 1
	file[3] = 1
	user.StoreFile("test", file)
	loadedFile, err3 = user.LoadFile("test")
	if err3 != nil {
		t.Error("Failed to upload and download", err3)
		return
	}
	if !reflect.DeepEqual(file, loadedFile) {
		t.Log("StoreFile overwrite failed. This is acceptable.")
		// Some implementations don't implement overwrite so this is not a fail.
	}

}

func TestStorageWithCorruptDatastore(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	fileNames := []string{"f1", "f2", "f3", "f4", "f5"}
	userNames := []string{"u1", "u2", "u3", "u4", "u5"}
	for i, offset := range []int{-4, -1, 0, 1, 7} {
		user, err1 := InitUser(userNames[i], "fubar")
		if err1 != nil {
			t.Error(err1)
			return
		}

		file := userlib.RandomBytes(userlib.AESBlockSize*7 - offset)
		user.StoreFile(fileNames[i], file)

		// Get user to check for userdata update.
		user, err2 := GetUser(userNames[i], "fubar")
		if err2 != nil {
			t.Error(err2)
			return
		}

		var keys []userlib.UUID
		var vals [][]byte
		for k, v := range datastore {
			keys = append(keys, k)
			vals = append(vals, v)
		}

		errored := false
		for k := range keys {
			datastore[keys[k]] = userlib.RandomBytes(len(vals[k]))
			_, err := user.LoadFile(fileNames[i])
			if err != nil {
				errored = true
			}
			datastore[keys[k]] = vals[k]
		}

		if !errored {
			t.Error("Corrupted datastore but no failed file load.")
		}
	}
}

func TestAppendBasic(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	fileNames := []string{"f1", "f2", "f3", "f4", "f5"}
	userNames := []string{"u1", "u2", "u3", "u4", "u5"}
	for i, offset1 := range []int{-4, -1, 0, 1, 7} {
		for _, offset2 := range []int{0, 1, 7} {
			userlib.DatastoreClear()
			userlib.KeystoreClear()

			user, err0 := InitUser(userNames[i], "fubar")
			if err0 != nil {
				t.Error(err0)
				return
			}

			file := userlib.RandomBytes(userlib.AESBlockSize - offset1)
			toAppend := userlib.RandomBytes(userlib.AESBlockSize * offset2)

			user.StoreFile(fileNames[i], file)
			err1 := user.AppendFile(fileNames[i], toAppend)
			if err1 != nil {
				t.Error(err1)
				return
			}

			// Get user to check for userdata update.
			user, err2 := GetUser(userNames[i], "fubar")
			if err2 != nil {
				t.Error(err2)
				return
			}

			loadedFile, err3 := user.LoadFile(fileNames[i])
			if err3 != nil {
				t.Error(err3)
				return
			}
			refAppend := append(file, toAppend...)
			if !reflect.DeepEqual(refAppend, loadedFile) {
				t.Error("Loaded (appended) file is not the same as reference\n",
					refAppend, "\n", loadedFile)
				return
			}
		}
	}

	user, err2 := GetUser("u5", "fubar")
	if err2 != nil {
		t.Error(err2)
		return
	}
	err1 := user.AppendFile("wrong", []byte{0, 0})
	if err1 == nil {
		t.Error("Appended to a file that does not exist.")
		return
	}
	file, err := user.LoadFile("wrong")
	if err != nil || file != nil {
		t.Error("Loaded a file that does not exist.")
		return
	}
}

func TestAppendMultipleFiles(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	fileNames := []string{"f1", "f2", "f3", "f4", "f5"}
	userNames := []string{"u1", "u2", "u3", "u4", "u5"}
	for i, offset1 := range []int{-4, -1, 0, 1, 7} {
		user, err0 := InitUser(userNames[i], "fubar")
		if err0 != nil {
			t.Error(err0)
			return
		}

		file := userlib.RandomBytes(userlib.AESBlockSize - offset1)
		toAppend := userlib.RandomBytes(userlib.AESBlockSize)

		user.StoreFile(fileNames[i], file)
		err1 := user.AppendFile(fileNames[i], toAppend)
		if err1 != nil {
			t.Error(err1)
			return
		}

		// Get user to check for userdata update.
		user, err2 := GetUser(userNames[i], "fubar")
		if err2 != nil {
			t.Error(err2)
			return
		}

		loadedFile, err3 := user.LoadFile(fileNames[i])
		if err3 != nil {
			t.Error(err3)
			return
		}
		refAppend := append(file, toAppend...)
		if !reflect.DeepEqual(refAppend, loadedFile) {
			t.Error("Loaded (appended) file is not the same as reference\n",
				refAppend, "\n", loadedFile)
			return
		}
	}
}

func TestAppendWithCorruptDatastore(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	user, err0 := InitUser("bob", "fubar")
	if err0 != nil {
		t.Error(err0)
		return
	}

	file := userlib.RandomBytes(userlib.AESBlockSize)
	toAppend := userlib.RandomBytes(userlib.AESBlockSize)

	user.StoreFile("test", file)

	var keys []userlib.UUID
	var vals [][]byte
	for k, v := range datastore {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	userlib.DatastoreSet(keys[0], vals[1])
	for i := 1; i < len(keys); i++ {
		userlib.DatastoreSet(keys[i], vals[0])
	}

	err1 := user.AppendFile("test", toAppend)
	if err1 == nil {
		t.Error("Successful append on a corrupted file")
		return
	}

	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore = userlib.DatastoreGetMap()

	fileNames := []string{"f1", "f2", "f3", "f4", "f5"}
	userNames := []string{"u1", "u2", "u3", "u4", "u5"}
	for i, offset := range []int{-4, -1, 0, 1, 7} {
		user, err1 := InitUser(userNames[i], "fubar")
		if err1 != nil {
			t.Error(err1)
			return
		}

		file := userlib.RandomBytes(userlib.AESBlockSize*7 - offset)
		toAppend := userlib.RandomBytes(userlib.AESBlockSize)
		user.StoreFile(fileNames[i], file)
		err1 = user.AppendFile(fileNames[i], toAppend)
		if err1 != nil {
			t.Error("Failed to append")
			return
		}

		loadedFile, err3 := user.LoadFile(fileNames[i])
		if err3 != nil {
			t.Error(err3)
			return
		}
		refAppend := append(file, toAppend...)
		if !reflect.DeepEqual(refAppend, loadedFile) {
			t.Error("Loaded (appended) file is not the same as reference\n",
				refAppend, "\n", loadedFile)
			return
		}

		var keys []userlib.UUID
		var vals [][]byte
		for k, v := range datastore {
			keys = append(keys, k)
			vals = append(vals, v)
		}

		errored := false
		for k := range keys {
			datastore[keys[k]] = userlib.RandomBytes(len(vals[k]))
			loadedFile, err3 = user.LoadFile(fileNames[i])
			if err3 != nil {
				errored = true
			}
			datastore[keys[k]] = vals[k]
		}

		if !errored {
			t.Error("Corrupted datastore but no failed file load.")
		}
	}
}

func TestAppend(t *testing.T) {
	file := userlib.RandomBytes(userlib.AESBlockSize * 500)
	randVal := userlib.RandomBytes(1)
	bob, err := InitUser("bob", "fubar")
	if bob == nil || err != nil {
		t.Error(err)
		return
	}
	bob.StoreFile("test", file)

	var sum time.Duration
	for i := 0; i < 5000; i++ {
		start := time.Now()
		_, err := bob.LoadFile("test")
		if err != nil {
			t.Error(err)
		}
		sum += time.Since(start)
	}
	avg := sum / 5000
	t.Log("average for loading is :", avg)

	sum = time.Since(time.Now())
	for i := 0; i < 5000; i++ {
		var keys []userlib.UUID
		var vals [][]byte
		for k, v := range userlib.DatastoreGetMap() {
			keys = append(keys, k)
			vals = append(vals, v)
		}

		start := time.Now()
		err := bob.AppendFile("test", randVal)
		if err != nil {
			t.Error(err)
		}
		sum += time.Since(start)

		userlib.DatastoreClear()
		for i := range keys {
			userlib.DatastoreSet(keys[i], vals[i])
		}
	}
	avg2 := sum / 5000
	t.Log("average for appending is :", avg2)

	if avg2 >= avg/2 {
		t.Error("Inefficient append")
	}
}

func TestShareBasic(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	file := userlib.RandomBytes(userlib.AESBlockSize)
	u.StoreFile("file1", file)

	var v2 []byte
	var magic_string string

	v, err := u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to load file1", err)
		return
	}
	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(file, v2) {
		t.Error("Shared file is not the same", file, v2)
		return
	}

	err = u.RevokeFile("file1")
	if err != nil {
		t.Error("Error revoking access", err)
		return
	}
	toAppend := []byte("sleepy")
	err = u.AppendFile("file1", toAppend)
	if err != nil {
		t.Error("Error appending", err)
	}
	v3, err := u.LoadFile("file1")
	v2, err = u2.LoadFile("file2")
	if reflect.DeepEqual(v3, v2) {
		t.Error("Bob should no longer have access or updates to file2", v, v2, v3)
		return
	}

}

func TestShareCorruptMagicString(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	alice, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice", err)
		return
	}
	bob, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	file := userlib.RandomBytes(userlib.AESBlockSize)
	alice.StoreFile("file1", file)

	var magic_string string

	_, err = alice.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = alice.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	var msg string

	// Slicing
	for i := 1; i < len(magic_string); i += 5 {
		msg = magic_string[:i]
		err = bob.ReceiveFile("file2", "alice", msg)
		if err == nil {
			t.Error("Manipulation of magic string should've errored")
			return
		}
		msg = magic_string[:(i/2)] + magic_string[(len(magic_string)-(i/2)):]
		err = bob.ReceiveFile("file2", "alice", msg)
		if err == nil {
			t.Error("Manipulation of magic string should've errored")
			return
		}
		msg = magic_string[:i] + string(userlib.RandomBytes(i)) + magic_string[i:]
		err = bob.ReceiveFile("file2", "alice", msg)
		if err == nil {
			t.Error("Manipulation of magic string should've errored")
			return
		}
	}

	// Random sharing string that isn't magic_string
	msg = string(userlib.RandomBytes(len(magic_string)))
	for msg == magic_string {
		msg = string(userlib.RandomBytes(len(magic_string)))
	}
	err = bob.ReceiveFile("file2", "alice", msg)
	if err == nil {
		t.Error("Random magic string should've errored")
		return
	}

	// Passing in empty sharing string
	msg = ""
	err = bob.ReceiveFile("file2", "alice", msg)
	if err == nil {
		t.Error("Empty magic string should've errored")
		return
	}
}

func TestShareUsernameMixup(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	alice, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice", err)
		return
	}
	bob, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	carol, err3 := InitUser("carol", "yesterday")
	if err3 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	file1 := userlib.RandomBytes(userlib.AESBlockSize)
	alice.StoreFile("file1", file1)

	var v, v2 []byte
	var magic_string string

	v, err = alice.LoadFile("file1")
	magic_string, err = alice.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	// Bob can get it
	err = bob.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	// Carol can't
	err = carol.ReceiveFile("file2", "alice", magic_string)
	if err == nil {
		t.Error("Failed to error with wrong recipient", err)
		return
	}

	// Bob got it
	v2, err = bob.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

	// Carol what are you doing? You don't have access!
	_, err = carol.LoadFile("file2")
	if err != nil {
		t.Error("Should error since Carol shouldn't have this filename", err)
		return
	}

	// Typo in sender name
	err = bob.ReceiveFile("file2", "alLice", magic_string)
	if err == nil {
		t.Error("Typo in sender username", err)
		return
	}

	// Typo in recipient name.
	magic_string, err = alice.ShareFile("file1", "boob")
	if err == nil {
		t.Error("Can't share with a nonexistent user.", err)
		return
	}
	err = bob.ReceiveFile("file2", "alice", magic_string)
	if err == nil {
		t.Error("Bob shouldn't have received access", err)
		return
	}
}

func TestShareFilenameMixup(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	alice, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice", err)
		return
	}
	bob, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	carol, err3 := InitUser("carol", "yesterday")
	if err3 != nil {
		t.Error("Failed to initialize bob", err3)
		return
	}

	file1 := userlib.RandomBytes(userlib.AESBlockSize)
	alice.StoreFile("file1", file1)

	file2 := userlib.RandomBytes(userlib.AESBlockSize * 2)
	bob.StoreFile("file2", file2)

	file3 := userlib.RandomBytes(userlib.AESBlockSize * 3)
	carol.StoreFile("file3", file3)

	_, err = alice.ShareFile("file2", "bob")
	if err == nil {
		t.Error("Can't share a file you don't have")
		return
	}
	_, err = alice.ShareFile("file3", "bob")
	if err == nil {
		t.Error("Can't share a file you don't have")
		return
	}
	_, err = bob.ShareFile("file1", "alice")
	if err == nil {
		t.Error("Can't share a file you don't have")
		return
	}
	_, err = bob.ShareFile("file3", "alice")
	if err == nil {
		t.Error("Can't share a file you don't have")
		return
	}
	_, err = carol.ShareFile("file1", "alice")
	if err == nil {
		t.Error("Can't share a file you don't have")
		return
	}

	var magic_string string
	magic_string, err = alice.ShareFile("file1", "carol")
	if err != nil {
		t.Error("Failed to share", err)
		return
	}
	err = carol.ReceiveFile("area51", "alice", magic_string)
	if err != nil {
		t.Error("Sharing failed")
		return
	}
}

func TestShareIntegration(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	alice, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice", err)
		return
	}
	bob, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	carol, err3 := InitUser("carol", "23lkrDSadfj?sdlkfj3")
	if err3 != nil {
		t.Error("Failed to initialize bob", err3)
		return
	}

	file := []byte("AS A MEMBER OF THE GREEK COMMUNITY AND PART OF ONE OF THESE ORGANIZATIONS, " +
		"THIS IS HIGHLY OFFENSIVE. SORORITIES AT UC BERKELEY MAKE IT THEIR GOAL TO GIVE " +
		"WOMEN A PLACE TO FEEL COMFORTABLE AS WELL AS BETTER THE COMMUNITY. COMPARING " +
		"SPECIFIC HOUSES TO CHARACTERS FROM A MOVIE ABOUT BULLYING IS ABSURD AND BEYON" +
		"D INACCURATE. MAKING THE CLAIM THAT SORORITIES ARE CLIQUES IS DEMEANING THE S" +
		"ISTERHOOD AND VALUES THAT THEY ARE FOUNDED ON. THIS CLEARLY IS A STAB AT A CO" +
		"MMUNITY ON CAMPUS THAT DOES NOTHING BUT SUPPORT THE REST OF THE STUDENT BODY.")

	alice.StoreFile("F", file)
	var ms1, ms2 string
	ms1, err = alice.ShareFile("F", "bob")
	if err != nil {
		t.Error(err)
		return
	}
	err = bob.ReceiveFile("memes", "alice", ms1)
	if err != nil {
		t.Error(err)
		return
	}
	ms2, err = bob.ShareFile("memes", "carol")
	if err != nil {
		t.Error(err)
		return
	}
	err = carol.ReceiveFile("lol", "alice", ms2)
	if err == nil {
		t.Error("Sender was bob, not alice. Should've errored")
		return
	}
	err = carol.ReceiveFile("lol", "bob", ms2)
	if err != nil {
		t.Error("Sharing failed", err)
		return
	}

	toAppend := []byte("UC Berkeley loses ranking due to misreporting of donations.")
	newFile := append(file, toAppend...)
	err = carol.AppendFile("lol", toAppend)
	if err != nil {
		t.Error(err)
		return
	}
	var a, b, c []byte
	a, err = alice.LoadFile("F")
	if err != nil || !reflect.DeepEqual(a, newFile) {
		t.Error("Updates not loaded")
		return
	}
	b, err = bob.LoadFile("memes")
	if err != nil || !reflect.DeepEqual(b, newFile) {
		t.Error("Updates not loaded")
		return
	}
	c, err = carol.LoadFile("lol")
	if err != nil || !reflect.DeepEqual(c, newFile) {
		t.Error("Updates not loaded")
		return
	}

	err = alice.RevokeFile("F")
	if err != nil {
		t.Error(err)
		return
	}
	toAppend = []byte("UCBMFET")
	newFile = append(newFile, toAppend...)
	err = bob.AppendFile("memes", toAppend)
	if err == nil {
		t.Error("Bob should not have been able to update the file")
		return
	}
	err = carol.AppendFile("lol", toAppend)
	if err == nil {
		t.Error("Carol should not have been able to update the file")
		return
	}

	err = alice.AppendFile("F", toAppend)
	if err != nil {
		t.Error(err)
		return
	}

	b, _ = bob.LoadFile("memes")
	if reflect.DeepEqual(b, newFile) {
		t.Error("Bob should not be able to see updates from Alice's revoke call")
		return
	}
	c, _ = carol.LoadFile("lol")
	if reflect.DeepEqual(c, newFile) {
		t.Error("Carol should not be able to see updates from Alice's revoke call")
		return
	}

	err = bob.ReceiveFile("me", "alice", ms1)
	if err != nil {
		t.Error(err)
		return
	}
	c, _ = bob.LoadFile("me")
	if reflect.DeepEqual(c, newFile) {
		t.Error("Bob should not be able to regain access")
		return
	}
	err = carol.ReceiveFile("lol", "bob", ms2)
	if err != nil {
		t.Error(err)
		return
	}
	c, _ = carol.LoadFile("lol")
	if reflect.DeepEqual(c, newFile) {
		t.Error("Carol should not be able to regain access")
		return
	}
}

func TestStuffAfterRevoke(t *testing.T) {
	userlib.SetDebugStatus(false)
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	datastore := userlib.DatastoreGetMap()
	keystore := userlib.KeystoreGetMap()
	_, _ = datastore, keystore

	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize alice", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	file := userlib.RandomBytes(userlib.AESBlockSize)
	u.StoreFile("file1", file)

	magic_string, err := u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err := u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(file, v2) {
		t.Error("Shared file is not the same", file, v2)
		return
	}

	err = u.RevokeFile("file1")
	if err != nil {
		t.Error("Revoke failed")
	}

	file2, err := u2.LoadFile("file2")
	if err != nil || reflect.DeepEqual(file2, file) {
		t.Error("Loaded a file that was revoked")
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	toAppend := userlib.RandomBytes(userlib.AESBlockSize)

	err = u.AppendFile("file1", toAppend)
	if err != nil {
		t.Error(err)
		return
	}

	file2, err = u2.LoadFile("file2")
	if err != nil || file2 == nil {
		t.Error("Failed to load a shared file")
		return
	}
	if !reflect.DeepEqual(file2, append(file, toAppend...)) {
		t.Error("Receiver cannot view edits to shared file.")
	}

	err = u.RevokeFile("file1")
	if err != nil {
		t.Error("Revoke failed")
	}

	err = u2.AppendFile("file2", toAppend)
	if err == nil {
		t.Error("Able to append to a revoked file")
		return
	}
}
