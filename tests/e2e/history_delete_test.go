package e2e

import (
	"testing"

	pubnub "github.com/pubnub/go"
	"github.com/stretchr/testify/assert"
)

func TestHistoryDeleteNotStubbed(t *testing.T) {
	assert := assert.New(t)

	ch := randomized("h-ch")
	pn := pubnub.NewPubNub(config)

	_, _, err := pn.DeleteMessages().
		Channel(ch).
		Execute()

	assert.Nil(err)
}

func TestHistoryDeleteMissingChannelError(t *testing.T) {
	assert := assert.New(t)

	pn := pubnub.NewPubNub(config)

	res, _, err := pn.DeleteMessages().
		Channel("").
		Execute()

	assert.Nil(res)
	assert.Contains(err.Error(), "Missing Channel")
}

func TestHistoryDeleteSuperCall(t *testing.T) {
	assert := assert.New(t)

	config := pamConfigCopy()

	// Not allowed characters: /?#,
	validCharacters := "-._~:[]@!$&'()*+;=`|"

	config.Uuid = validCharacters
	config.AuthKey = validCharacters

	pn := pubnub.NewPubNub(config)

	_, _, err := pn.DeleteMessages().
		Channel(validCharacters).
		Start(int64(123)).
		End(int64(456)).
		Execute()

	assert.Nil(err)
}
