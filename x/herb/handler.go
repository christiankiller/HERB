package herb

import (
	"fmt"

	"github.com/corestario/HERB/x/herb/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "herb" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSetCiphertextShare:
			return handleMsgSetCiphertextShare(ctx, &keeper, msg)
		case MsgSetDecryptionShare:
			return handleMsgSetDecryptionShare(ctx, &keeper, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized herb Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSetCiphertextShare(ctx sdk.Context, keeper *Keeper, msg types.MsgSetCiphertextShare) sdk.Result {
	ctShare, err := msg.CiphertextShare.Deserialize()
	if err != nil {
		return sdk.ErrUnknownRequest(fmt.Sprintf("can't deserialize ciphertext share: %v", err)).Result()
	}
	if err := keeper.SetCiphertext(ctx, ctShare); err != nil {
		return err.Result()
	}
	return sdk.Result{}
}

func handleMsgSetDecryptionShare(ctx sdk.Context, keeper *Keeper, msg types.MsgSetDecryptionShare) sdk.Result {
	decryptionShare, err := msg.DecryptionShare.Deserialize()
	if err != nil {
		return sdk.ErrUnknownRequest(fmt.Sprintf("can't deserialize decryption share: %v", err)).Result()
	}
	if err := keeper.SetDecryptionShare(ctx, decryptionShare); err != nil {
		return err.Result()
	}
	return sdk.Result{}
}
