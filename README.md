# Ergo-golang

## Documentation

### Send transaction only with sk

```golang
import (
	"github.com/zhiganov-andrew/ergo-golang/pkg/transaction"
)

func main() {
	transaction.SendTransaction(recipient, amount, fee, sk, testnet)
}
```
