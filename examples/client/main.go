package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/kinecosystem/agora-common/kin"

	commonpb "github.com/kinecosystem/agora-api/genproto/common/v3"

	"github.com/kinecosystem/kin-go/client"
)

var (
	senderSeed  = flag.String("sender", "", "Sender seed")
	destAddress = flag.String("dest", "", "Destination address")
)

func main() {
	flag.Parse()

	sender, err := kin.PrivateKeyFromString(*senderSeed)
	if err != nil {
		log.Fatal(err)
	}
	dest, err := kin.PublicKeyFromString(*destAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the SDK using AppIndex 2, the test app.
	c, err := client.New(client.EnvironmentTest, client.WithAppIndex(1), client.WithMaxRetries(0))
	if err != nil {
		log.Fatal(err)
	}

	// Create a new account
	err = c.CreateAccount(context.Background(), sender)
	if err != nil {
		log.Fatal(err)
	}

	// Payment with no invoicing.
	txHash, err := c.SubmitPayment(context.Background(), client.Payment{
		Sender:      sender,
		Destination: dest,
		Type:        kin.TransactionTypeP2P,
		Quarks:      kin.MustToQuarks("1"),
	}, client.WithAccountResolution(client.AccountResolutionPreferred))
	fmt.Printf("Hash: %x, err: %v\n", txHash, err)

	// Payment with an old style memo
	txHash, err = c.SubmitPayment(context.Background(), client.Payment{
		Sender:      sender,
		Destination: dest,
		Type:        kin.TransactionTypeSpend,
		Quarks:      kin.MustToQuarks("1.0"),
		Memo:        "1-test",
	})

	// Payment with an invoice
	txHash, err = c.SubmitPayment(context.Background(), client.Payment{
		Sender:      sender,
		Destination: dest,
		Type:        kin.TransactionTypeSpend,
		Quarks:      kin.MustToQuarks("1.0"),
		Invoice: &commonpb.Invoice{
			Items: []*commonpb.Invoice_LineItem{
				{
					Title:       "TestPayment",
					Description: "Optional desc about the payment",
					Amount:      kin.MustToQuarks("1.0"),
					Sku:         []byte("some opaque sky"),
				},
			},
		},
	})
	fmt.Printf("Hash: %x, err: %v\n", txHash, err)

	// Payment with dedupe
	dedupeID := uuid.New()
	payment := client.Payment{
		Sender:      sender,
		Destination: dest,
		Type:        kin.TransactionTypeP2P,
		Quarks:      kin.MustToQuarks("1"),
		DedupeID:    dedupeID[:],
	}
	txHash, err = c.SubmitPayment(context.Background(), payment)
	if err != nil {
		// Safe to retry since DedupeID was set
		txHash, err = c.SubmitPayment(context.Background(), payment)
	}
	fmt.Printf("Hash: %x, err: %v\n", txHash, err)

	// Earn batch with an old style memo + dedupe
	dedupeID = uuid.New()
	batch := client.EarnBatch{
		Sender: sender,
		Memo:   "1-test",
		Earns: []client.Earn{
			{
				Destination: dest,
				Quarks:      kin.MustToQuarks("1.0"),
			},
			{
				Destination: dest,
				Quarks:      kin.MustToQuarks("1.0"),
			},
		},
		DedupeID: dedupeID[:],
	}
	result, err := c.SubmitEarnBatch(context.Background(), batch)
	if err != nil {
		// Safe to retry since DedupeID was set
		result, err = c.SubmitEarnBatch(context.Background(), batch)
		if err != nil {
			log.Fatal(err)
		}
	}

	if result.TxError != nil {
		fmt.Printf("\tHash: %x, Error: %v\n", result.TxID, result.TxError)

		if result.EarnErrors != nil {
			for _, ee := range result.EarnErrors {
				fmt.Printf("\tEarnIndex: %x, Error: %v\n", ee.EarnIndex, ee.Error)
			}
		}
	} else {
		fmt.Printf("\tHash: %x, succeeded\n", result.TxID)
	}
}
