package tests

import (
	"github.com/troian/surgemq/tests/mqtt/test1"
	"github.com/troian/surgemq/tests/mqtt/test10"
	"github.com/troian/surgemq/tests/mqtt/test11"
	"github.com/troian/surgemq/tests/mqtt/test12"
	"github.com/troian/surgemq/tests/mqtt/test13"
	"github.com/troian/surgemq/tests/mqtt/test2"
	"github.com/troian/surgemq/tests/mqtt/test3"
	"github.com/troian/surgemq/tests/mqtt/test4"
	"github.com/troian/surgemq/tests/mqtt/test5"
	"github.com/troian/surgemq/tests/mqtt/test6"
	"github.com/troian/surgemq/tests/mqtt/test7"
	"github.com/troian/surgemq/tests/mqtt/test8"
	testTypes "github.com/troian/surgemq/tests/types"
)

func init() {
	testList = []testTypes.Provider{
		test1.New(),
		test2.New(),
		test3.New(),
		test4.New(),
		test5.New(),
		test6.New(),
		test7.New(),
		test8.New(),
		test10.New(),
		test11.New(),
		test12.New(),
		test13.New(),
	}
}
