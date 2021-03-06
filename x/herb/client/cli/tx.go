package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/corestario/HERB/x/herb/elgamal"
	"github.com/corestario/HERB/x/herb/types"

	"github.com/spf13/cobra"

	"go.dedis.ch/kyber/v3/share"
	kyberenc "go.dedis.ch/kyber/v3/util/encoding"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	herbTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "HERB transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	herbTxCmd.AddCommand(client.PostCommands(
		GetCmdSetCiphertextShare(cdc),
		GetCmdSetDecryptionShare(cdc),
	)...)

	return herbTxCmd
}

// GetCmdSetCiphertext implements send ciphertext share transaction command.
func GetCmdSetCiphertextShare(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ct-share [commonPubKey]",
		Short: "send random ciphertext share",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			group := types.P256
			pubKey, err := kyberenc.StringHexToPoint(group, args[0])
			if err != nil {
				return fmt.Errorf("failed to decode common public key: %v", err)
			}

			ct, ceproof, err := elgamal.RandomCiphertext(group, pubKey)
			if err != nil {
				return fmt.Errorf("failed to create random ciphertext: %v", err)
			}

			sender := cliCtx.GetFromAddress()
			ctShare := types.CiphertextShare{Ciphertext: ct, CEproof: ceproof, EntropyProvider: sender}
			ctShareJSON, err := types.NewCiphertextShareJSON(&ctShare)
			if err != nil {
				return err
			}
			msg := types.NewMsgSetCiphertextShare(*ctShareJSON, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetDecryptionShare implements send decryption share transaction command.
func GetCmdSetDecryptionShare(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "decrypt [privateKey] [ID]",
		Short: "Send a decryption share of the aggregated ciphertext",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			//Getting aggregated ciphertext
			params := types.NewQueryByRound(-1) //-1 for the current round
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			ctShareBytes, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRouter, types.QueryAggregatedCt), bz)
			if err != nil {
				return err
			}

			var ctJSON types.QueryAggregatedCtRes
			cdc.MustUnmarshalJSON(ctShareBytes, &ctJSON)
			aggregatedCt, err := ctJSON.CiphertextJSON.Deserialize(types.P256)
			if err != nil {
				return err
			}

			//decrypting ciphertext
			group := types.P256
			privKey, err := kyberenc.StringHexToScalar(group, args[0])
			if err != nil {
				return fmt.Errorf("failed to decode private key: %v", err)
			}

			id, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("id %s not a valid int, please input a valid id", args[1])
			}

			sharePoint, proof, err := elgamal.CreateDecShare(group, *aggregatedCt, privKey)
			if err != nil {
				return err
			}

			decryptionShare := &types.DecryptionShare{
				DecShare:      share.PubShare{I: int(id), V: sharePoint},
				DLEQproof:      proof,
				KeyHolderAddr: cliCtx.GetFromAddress(),
			}

			decryptionShareJSON, err := types.NewDecryptionShareJSON(decryptionShare)
			if err != nil {
				return err
			}
			msg := types.NewMsgSetDecryptionShare(decryptionShareJSON, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			txBldr, err = utils.EnrichWithGas(txBldr, cliCtx, []sdk.Msg{msg})
			if err != nil {
				return err
			}
			txBldr = txBldr.WithGas(5 * txBldr.Gas())
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
