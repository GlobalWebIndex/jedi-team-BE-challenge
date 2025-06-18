package vectordb

import (
	"context"
	"fmt"
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
	"github.com/loukaspe/jedi-team-challenge/pkg/helpers"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

type PineconeVectorDB struct {
	client            *pinecone.Client
	index             string
	topKResultsNumber int
	threshold         float32
}

func NewPineconeVectorDB(threshold float32, topKResultsNumber int, index string, client *pinecone.Client) *PineconeVectorDB {
	return &PineconeVectorDB{
		topKResultsNumber: topKResultsNumber,
		index:             index,
		client:            client,
		threshold:         threshold,
	}
}

func (db *PineconeVectorDB) StoreEmbeddings(ctx context.Context, embeddings []*domain.Embeddings, extraMetadata map[string]interface{}) (int, error) {
	idx, err := db.client.DescribeIndex(ctx, db.index)
	if err != nil {
		return 0, err
	}

	idxConnection, err := db.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return 0, err
	}

	vectors := make([]*pinecone.Vector, len(embeddings))
	for i, embedding := range embeddings {
		id := fmt.Sprintf("doc1-chunk-%d", i)

		md, err := structpb.NewStruct(map[string]interface{}{
			"text": embeddings[i].Text,
		})
		if err != nil {
			return 0, err
		}

		if len(extraMetadata) > 0 {
			for key, value := range extraMetadata {
				md.Fields[key], err = structpb.NewValue(value)
				if err != nil {
					return 0, err
				}
			}
		}

		vectorToFloat32 := helpers.Float64ToFloat32(embedding.Embeddings)

		vectors[i] = &pinecone.Vector{
			Id:       id,
			Values:   &vectorToFloat32,
			Metadata: md,
		}
	}

	count, err := idxConnection.UpsertVectors(ctx, vectors)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (db *PineconeVectorDB) SemanticSearch(ctx context.Context, embeddings []float32) ([]string, error) {
	idx, err := db.client.DescribeIndex(ctx, db.index)
	if err != nil {
		return []string{}, err
	}

	idxConnection, err := db.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return []string{}, err
	}

	res, err := idxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          embeddings,
		TopK:            uint32(db.topKResultsNumber),
		IncludeValues:   false,
		IncludeMetadata: true,
	})

	var contextTexts []string
	for _, match := range res.Matches {
		text := match.Vector.Metadata.String()
		score := match.Score

		if score >= db.threshold {
			contextTexts = append(contextTexts, text)
		}
	}

	return contextTexts, nil
}
