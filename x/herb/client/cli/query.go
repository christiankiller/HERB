package cli

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/dgamingfoundation/HERB/x/herb/elgamal"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/dgamingfoundation/HERB/x/herb/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	herbQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the herb module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	herbQueryCmd.AddCommand(client.GetCommands(
		GetCmdAggregatedCiphertext(storeKey, cdc),
		GetCmdAllCiphertexts(storeKey, cdc),
		GetCmdAllDecryptionShares(storeKey, cdc),
		GetCmdRandom(storeKey, cdc),
		GetCmdCurrentRound(storeKey, cdc),
		GetCmdKeyHoldersNumber(storeKey, cdc),
		GetCmdRoundStage(storeKey, cdc),
	)...)

	return herbQueryCmd
}

// GetCmdAggregatedCiphertext implements the query aggregated ciphertext command.
func GetCmdAggregatedCiphertext(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "aggregated-ct [round]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Query aggregated elgamal ciphertext for the given round",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var round int64
			if len(args) > 0 {
				parsedRound, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("round %s not a valid uint, please input a valid round", args[0])
				}
				round = int64(parsedRound)
			} else {
				round = -1
			}

			params := types.NewQueryByRound(round)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryAggregatedCt), bz)
			if err != nil {
				return err
			}
			var outJSON elgamal.CiphertextJSON
			cdc.MustUnmarshalJSON(res, &outJSON)
			out, err := outJSON.Deserialize(types.P256)
			if err != nil {
				return err
			}
			fmt.Println(out.String())
			return nil
			//return cliCtx.PrintOutput(outJSON)
		},
	}
}

// GetCmdAllCiphertexts implements the query of all ciphertexts command.
func GetCmdAllCiphertexts(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "all-ct [round]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Query all elgamal ciphertexts for the given round",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var round int64
			if len(args) > 0 {
				parsedRound, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("round %s not a valid uint, please input a valid round", args[0])
				}
				round = int64(parsedRound)
			} else {
				round = -1
			}

			params := types.NewQueryByRound(round)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryAllCt), bz)
			if err != nil {
				return err
			}
			outJSON := make(map[string]*types.CiphertextPartJSON)
			cdc.MustUnmarshalJSON(res, &outJSON)
			out, err := types.CiphertextMapDeserialize(outJSON)
			if err != nil {
				return err
			}
			for _, ctPart := range out {
				fmt.Printf("Entropy provider address: %v \n Ciphertext: %v \n", ctPart.EntropyProvider.String(), ctPart.Ciphertext.String())
			}
			return nil
			//return cliCtx.PrintOutput(outJSON)
		},
	}
}

// GetCmdAllDecryptionShares implements the query of all decryption shares command.
func GetCmdAllDecryptionShares(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "all-shares [round]",
		Short: "Queries all decryption shares for the given round",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var round int64
			if len(args) > 0 {
				parsedRound, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("round %s not a valid uint, please input a valid round", args[0])
				}
				round = int64(parsedRound)
			} else {
				round = -1
			}

			params := types.NewQueryByRound(round)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryAllDescryptionShares), bz)
			if err != nil {
				return err
			}

			outJSON := make(map[string]*types.DecryptionShareJSON)
			cdc.MustUnmarshalJSON(res, &outJSON)
			out, err := types.DecryptionSharesMapDeserialize(outJSON)

			if err != nil {
				return err
			}
			for _, share := range out {
				fmt.Printf("Key holder address: %v \n Descryption Share: %v \n", share.KeyHolder.String(), share.DecShare.V.String())
			}
			return nil
		},
	}
}

// GetCmdRandom implements the query of the random number result.
func GetCmdRandom(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-random [round]",
		Short: "Queries the random number generation result",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var round int64
			if len(args) > 0 {
				parsedRound, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("round %s not a valid uint, please input a valid round", args[0])
				}
				round = int64(parsedRound)
			} else {
				round = -1
			}

			params := types.NewQueryByRound(round)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryRandom), bz)
			if err != nil {
				return err
			}

			fmt.Printf("random data: %v", res)

			return nil
		},
	}
}

// GetCmdKeyHoldersNumber is required for the testing purposes
func GetCmdKeyHoldersNumber(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "kh-number",
		Short: "returns key holders number",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryKeyHoldersNumber), nil)
			if err != nil {
				return err
			}

			n := binary.LittleEndian.Uint64(res)

			fmt.Printf("number of key holders: %v", n)

			return nil
		},
	}
}

func GetCmdCurrentRound(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "current-round",
		Short: "returns current generation round",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			roundBytes, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCurrentRound), nil)
			if err != nil {
				return err
			}

			round := binary.LittleEndian.Uint64(roundBytes)

			fmt.Print(round)

			return nil
		},
	}
}

func GetCmdRoundStage(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "stage [round]",
		Short: "returns rounds stage",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var round int64
			if len(args) > 0 {
				parsedRound, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("round %s not a valid uint, please input a valid round", args[0])
				}
				round = int64(parsedRound)
			} else {
				round = -1
			}

			params := types.NewQueryByRound(round)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			stageBytes, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryStage), bz)

			fmt.Print(string(stageBytes))

			return nil
		},
	}
}
