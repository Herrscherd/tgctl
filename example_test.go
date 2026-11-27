package tgctl_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Herrscherd/tgctl"
)

func ExampleClient_GetMe() {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		reply := `{"ok":true,"result":{"id":777,"is_bot":true,"first_name":"Herrscher","username":"herrscher_bot"}}`
		if _, err := fmt.Fprint(writer, reply); err != nil {
			fmt.Println(err)
		}
	}))
	defer server.Close()

	bot := tgctl.New("123456:token", tgctl.WithBaseURL(server.URL))
	me, err := bot.GetMe(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(me.Username)
	// Output: herrscher_bot
}
