# Ergo-golang

## Documentation

### Send transaction only with sk

```golang
import (
	"github.com/ergo-golang/pkg/transaction"
)

func main() {
	transaction.SendTransaction(recipient, amount, fee, sk, testnet)
}
```
