package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeBlock(t *testing.T) {
	//
	b := NewBlock(0, [32]byte{})
	//
	assert.Equal(t, b.nonce, 0)
	assert.Equal(t, b.previousHash, [32]byte{})
	//
	assert.NotEqual(t, b.nonce, 1)
	assert.NotEqual(t, b.previousHash, "Final hash")
}
