package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseMessageWithEmptyContent(t *testing.T) {
	content := ``
	msg := ParseMessage(content)

	assert.Nil(t, msg.Keys)
	assert.Nil(t, msg.Headers)
	assert.Equal(t, "", msg.Body)
}

func Test_ParseMessageWithSimpleBody(t *testing.T) {
	content := `hello world`
	msg := ParseMessage(content)

	assert.Nil(t, msg.Keys)
	assert.Nil(t, msg.Headers)
	assert.Equal(t, "hello world", msg.Body)
}

func Test_ParseMessageWithAllParts(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
foo
bar
--key
id:12345

hello:world
badheader
--header
foo
fighters`

	msg := ParseMessage(content)

	assert.NotNil(t, msg.Keys)
	assert.Equal(t, "foo\nbar", msg.Keys[0].Value)
	assert.NotNil(t, msg.Headers)
	assert.Equal(t, 2, len(msg.Headers))
	assert.Equal(t, "id", msg.Headers[0].Key)
	assert.Equal(t, "12345", msg.Headers[0].Value)
	assert.Equal(t, "hello", msg.Headers[1].Key)
	assert.Equal(t, "world", msg.Headers[1].Value)
	assert.Equal(t, "foo\nfighters", msg.Body)
}

func Test_ParseMessageOnlyKey(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
foo
bar
--key`

	msg := ParseMessage(content)

	assert.NotNil(t, msg.Keys)
	assert.Equal(t, "foo\nbar", msg.Keys[0].Value)
	assert.Nil(t, msg.Headers)
	assert.Empty(t, msg.Body)
}

func Test_ParseMessageNoKey(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
id:12345
--header
foo
fighters`

	msg := ParseMessage(content)

	assert.Nil(t, msg.Keys)
	assert.NotNil(t, msg.Headers)
	assert.Equal(t, 1, len(msg.Headers))
	assert.Equal(t, "id", msg.Headers[0].Key)
	assert.Equal(t, "12345", msg.Headers[0].Value)
	assert.Equal(t, "foo\nfighters", msg.Body)
}

func Test_ParseMessageEmptyKey(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
--key
id:12345
--header
foo
fighters`

	msg := ParseMessage(content)

	assert.Nil(t, msg.Keys)
	assert.NotNil(t, msg.Headers)
	assert.Equal(t, 1, len(msg.Headers))
	assert.Equal(t, "id", msg.Headers[0].Key)
	assert.Equal(t, "12345", msg.Headers[0].Value)
	assert.Equal(t, "foo\nfighters", msg.Body)
}

func Test_ParseMessageOnlyHeaders(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
id:12345
hello:world
badheader
--header`

	msg := ParseMessage(content)

	assert.Nil(t, msg.Keys)
	assert.NotNil(t, msg.Headers)
	assert.Equal(t, 2, len(msg.Headers))
	assert.Equal(t, "id", msg.Headers[0].Key)
	assert.Equal(t, "12345", msg.Headers[0].Value)
	assert.Equal(t, "hello", msg.Headers[1].Key)
	assert.Equal(t, "world", msg.Headers[1].Value)
	assert.Empty(t, msg.Body)
}

func Test_ParseMessageOnlyHeadersSpacedOut(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
  id  :  12345
hello: world
badheader
--header`

	msg := ParseMessage(content)

	assert.Nil(t, msg.Keys)
	assert.NotNil(t, msg.Headers)
	assert.Equal(t, 2, len(msg.Headers))
	assert.Equal(t, "id", msg.Headers[0].Key)
	assert.Equal(t, "12345", msg.Headers[0].Value)
	assert.Equal(t, "hello", msg.Headers[1].Key)
	assert.Equal(t, "world", msg.Headers[1].Value)
	assert.Empty(t, msg.Body)
}

func Test_ParseMessageNoHeaders(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
gozer
--key
foo
fighters`

	msg := ParseMessage(content)

	assert.NotNil(t, msg.Keys)
	assert.Equal(t, "gozer", msg.Keys[0].Value)
	assert.Nil(t, msg.Headers)
	assert.Equal(t, "foo\nfighters", msg.Body)
}

func Test_ParseMessageEmptyHeaders(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
gozer
--key
--header
foo
fighters`

	msg := ParseMessage(content)

	assert.NotNil(t, msg.Keys)
	assert.Equal(t, "gozer", msg.Keys[0].Value)
	assert.Nil(t, msg.Headers)
	assert.Equal(t, "foo\nfighters", msg.Body)
}

func Test_ParseMessageWithMultipleKeyParts(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
foo
bar
--key
anotherKey
--key
foo
fighters`

	msg := ParseMessage(content)

	assert.NotNil(t, msg.Keys)
	assert.Equal(t, "foo\nbar", msg.Keys[0].Value)
	assert.Nil(t, msg.Headers)
	assert.Equal(t, "anotherKey\n--key\nfoo\nfighters", msg.Body)
}

func Test_ParseMessageWithMultipleHeaderParts(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
gozer
--key
hello:world
badheader
--header
test:me
--header
foo fighters`

	msg := ParseMessage(content)

	assert.NotNil(t, msg.Keys)
	assert.Equal(t, "gozer", msg.Keys[0].Value)
	assert.NotNil(t, msg.Headers)
	assert.Equal(t, 1, len(msg.Headers))
	assert.Equal(t, "hello", msg.Headers[0].Key)
	assert.Equal(t, "world", msg.Headers[0].Value)
	assert.Equal(t, "test:me\n--header\nfoo fighters", msg.Body)
}

func Test_ParseMessageWithMultipleParts(t *testing.T) {
	// make a sample message file content over multiple lines
	content := `
gozer
--key
anotherKey
--key
id:12345
--header
hello:world
--header
foo fighters`

	msg := ParseMessage(content)

	assert.NotNil(t, msg.Keys)
	assert.Equal(t, "gozer", msg.Keys[0].Value)
	assert.NotNil(t, msg.Headers)
	assert.Equal(t, 1, len(msg.Headers))
	assert.Equal(t, "id", msg.Headers[0].Key)
	assert.Equal(t, "12345", msg.Headers[0].Value)
	assert.Equal(t, "hello:world\n--header\nfoo fighters", msg.Body)
}
