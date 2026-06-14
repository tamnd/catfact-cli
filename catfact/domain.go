package catfact

import (
	"context"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes catfact as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/catfact-cli/catfact"
//
// The same Domain also builds the standalone catfact binary (see cli.NewApp).
func init() { kit.Register(Domain{}) }

// Domain is the catfact driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "catfact",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "catfact",
			Short:  "Random cat facts from catfact.ninja",
			Long: `catfact fetches random cat facts from the public catfact.ninja API.
No login or API key required.`,
			Site: Host,
			Repo: "https://github.com/tamnd/catfact-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// facts: list random cat facts
	kit.Handle(app, kit.OpMeta{
		Name:    "facts",
		Group:   "read",
		List:    true,
		Summary: "List random cat facts",
	}, factsOp)

	// fact: fetch one random cat fact
	kit.Handle(app, kit.OpMeta{
		Name:    "fact",
		Group:   "read",
		Single:  true,
		Summary: "Fetch one random cat fact",
	}, factOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type factsInput struct {
	Limit  int           `kit:"flag,inherit" help:"number of facts to return (default 10, max 332)"`
	Page   int           `kit:"flag" default:"1" help:"page number"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type factInput struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func factsOp(ctx context.Context, in factsInput, emit func(Fact) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	page := in.Page
	if page <= 0 {
		page = 1
	}
	items, err := in.Client.Facts(ctx, limit, page)
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func factOp(ctx context.Context, in factInput, emit func(Fact) error) error {
	item, err := in.Client.Fact(ctx)
	if err != nil {
		return mapErr(err)
	}
	return emit(item)
}

// --- Resolver ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty catfact reference")
	}
	return "fact", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "fact":
		return "https://catfact.ninja/fact", nil
	default:
		return "", errs.Usage("catfact has no resource type %q", uriType)
	}
}

// mapErr converts a library error into the kit error kind.
func mapErr(err error) error {
	return err
}
