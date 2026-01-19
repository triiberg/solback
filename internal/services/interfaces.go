package services

import (
	"context"

	"solback/internal/models"
)

type SourceProvider interface {
	GetSources(ctx context.Context) ([]models.Source, error)
}

type LogWriter interface {
	CreateLog(ctx context.Context, action string, outcome string, message *string) error
}

type HtmlFetcher interface {
	Fetch(ctx context.Context, url string) (HtmlResult, error)
}

type OpenAiExtractor interface {
	ExtractZipLink(ctx context.Context, html string) (OpenAiResult, error)
}

type ZipDownloader interface {
	Download(ctx context.Context, link string, sourceURL string) (ZipResult, error)
}

type ZipProcessor interface {
	ExtractAuctionPayloads(ctx context.Context, zipBytes []byte) ([]AuctionPayload, error)
}

type AuctionParser interface {
	ParseAuctionResults(ctx context.Context, payload AuctionPayload) (AuctionResults, error)
}

type DataStorer interface {
	StoreAuctionResults(ctx context.Context, results AuctionResults) (int, error)
}
