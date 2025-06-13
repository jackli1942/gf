package do

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Department is the data object for department operations.
type Department struct {
	Id        uint
	ParentId  uint
	Name      string
	Sort      int
	Status    *int
	Remark    string
	CreatedAt *gtime.Time
	UpdatedAt *gtime.Time
}
```
