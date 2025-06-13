package do

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Position is the data object for position operations.
type Position struct {
	Id        uint
	Name      string
	Code      string
	Sort      int
	Status    int
	Remark    string
	CreatedAt *gtime.Time
	UpdatedAt *gtime.Time
}
```
