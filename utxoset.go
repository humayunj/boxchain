package main

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

type UTXOSet struct {
	Blockchain *Blockchain
}

const utxoBucket = "utxoset"

func (u UTXOSet) ReIndex() {
	db := u.Blockchain.db

	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil {
			log.Print(err)
		}
		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			panic(err)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(t *bolt.Tx) error {
		b := t.Bucket(bucketName)
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			err = b.Put(key, outs.Serialize())
			if err != nil {
				panic(err)
			}
		}
		return nil
	})
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.isLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return UTXOs
}

func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}
					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							panic(err)
						} else {
							err := b.Put(vin.Txid, updatedOuts.Serialize())
							if err != nil {
								panic(err)
							}
						}

					}
				}

				newOutputs := TXOutputs{}
				for _, out := range tx.Vout {
					newOutputs.Outputs = append(newOutputs.Outputs, out)
				}

				err := b.Put(tx.ID, newOutputs.Serialize())
				if err != nil {
					panic(err)
				}
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)

	accumulated := 0

	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txId := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.isLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txId] = append(unspentOutputs[txId], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return accumulated, unspentOutputs
}
