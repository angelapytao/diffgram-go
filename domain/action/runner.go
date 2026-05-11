package action

import (
	"context"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

// Runner is the contract every Action implementation satisfies.
// Implementers must keep Run idempotent: the same ActionRun may be re-delivered.
type Runner interface {
	Name() string
	Run(ctx context.Context, run *entity.ActionRun) error
}
