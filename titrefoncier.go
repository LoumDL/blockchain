package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Définition de la structure des Titres Fonciers
type TitreFoncier struct {
	Id         string `json:"id"`         // Identifiant unique du titre foncier
	Proprio    string `json:"proprio"`    // Nom du propriétaire
	NumTF      string `json:"numTF"`      // Numéro officiel du titre foncier
	Superficie int    `json:"superficie"` // Superficie du terrain en m²
	Document   string `json:"document"`   // Chemin du fichier NFS
	DocHash    string `json:"doc_hash"`   // Hash SHA-1 du document
}

//Définition du Smart Contract
type SmartContract struct {
	contractapi.Contract
}

// Fonction pour calculer le hash SHA-1 d'un document
func GenerateSHA1Hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erreur d'ouverture fichier: %v", err)
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("erreur de lecture fichier: %v", err)
	}

	hash := sha1.Sum(content)
	return hex.EncodeToString(hash[:]), nil
}

// Initialisation avec quelques Titres Fonciers
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	titres := []TitreFoncier{
		{Id: "TF003", Proprio: "Djiby Loum", NumTF: "123456", Superficie: 700, Document: "/mnt/shared_dir/tf003.pdf", DocHash: "a8472b5ec66cfcb5ba20ae4e6b23c8c7277457df"},
		{Id: "TF002", Proprio: "Ndeye Fatou Dabo", NumTF: "6543211", Superficie: 1000, Document: "/mnt/shared_dir/tf002.pdf", DocHash: "6add312cd1ea92f19e803ee463cd7a8edc5736a8"},
	}

	for _, titre := range titres {
		titreJSON, err := json.Marshal(titre)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(titre.Id, titreJSON)
		if err != nil {
			return fmt.Errorf("erreur d'enregistrement: %v", err)
		}
	}

	return nil
}

//Ajouter un nouveau Titre Foncier
func (s *SmartContract) AjouterTitreFoncier(ctx contractapi.TransactionContextInterface, id string, proprio string, numTF string, superficie int, document string) error {
	// Vérifier si l'ID existe déjà
	existant, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("erreur de récupération de l'état: %v", err)
	}
	if existant != nil {
		return fmt.Errorf("le titre foncier %s existe déjà", id)
	}

	// Générer le hash du document
	docHash, err := GenerateSHA1Hash(document)
	if err != nil {
		return fmt.Errorf("erreur de génération du hash: %v", err)
	}

	// Créer l'objet
	titre := TitreFoncier{
		Id:         id,
		Proprio:    proprio,
		NumTF:      numTF,
		Superficie: superficie,
		Document:   document,
		DocHash:    docHash,
	}

	// Convertir en JSON et enregistrer
	titreJSON, err := json.Marshal(titre)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, titreJSON)
}

// Lire un Titre Foncier
func (s *SmartContract) LireTitreFoncier(ctx contractapi.TransactionContextInterface, id string) (*TitreFoncier, error) {
	titreJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("erreur de lecture: %v", err)
	}
	if titreJSON == nil {
		return nil, fmt.Errorf("titre foncier %s non trouvé", id)
	}

	var titre TitreFoncier
	err = json.Unmarshal(titreJSON, &titre)
	if err != nil {
		return nil, err
	}

	return &titre, nil
}

//Modifier un Titre Foncier (ex: mise à jour du propriétaire)
func (s *SmartContract) ModifierProprietaire(ctx contractapi.TransactionContextInterface, id string, nouveauProprio string) error {
	titre, err := s.LireTitreFoncier(ctx, id)
	if err != nil {
		return err
	}

	titre.Proprio = nouveauProprio

	titreJSON, err := json.Marshal(titre)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, titreJSON)
}

// Supprimer un Titre Foncier
func (s *SmartContract) SupprimerTitreFoncier(ctx contractapi.TransactionContextInterface, id string) error {
	existant, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression: %v", err)
	}
	if existant == nil {
		return fmt.Errorf("titre foncier %s introuvable", id)
	}

	return ctx.GetStub().DelState(id)
}

// Lister tous les Titres Fonciers
func (s *SmartContract) GetAllTitresFonciers(ctx contractapi.TransactionContextInterface) ([]*TitreFoncier, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var titres []*TitreFoncier
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var titre TitreFoncier
		err = json.Unmarshal(queryResponse.Value, &titre)
		if err != nil {
			return nil, err
		}
		titres = append(titres, &titre)
	}

	return titres, nil
}

func main() {
	titreChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Erreur création chaincode: %v", err)
	}

	if err := titreChaincode.Start(); err != nil {
		log.Panicf("Erreur démarrage chaincode: %v", err)
	}
}

