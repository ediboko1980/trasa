package accessproxytest

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/seknox/trasa/server/accessproxy/sshproxy"
	"github.com/seknox/trasa/server/models"
	"github.com/seknox/trasa/tests/server/testutils"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConnectNewSSH(t *testing.T) {

	type args struct {
		params models.ConnectionParams
	}

	tests := []struct {
		name       string
		args       args
		wantStatus bool
		wantErrMsg string
	}{
		{
			name: "should fail when hostname is incorrect",
			args: args{models.ConnectionParams{
				TotpCode:  testutils.GetTotpCode(testutils.MocktotpSEC),
				TfaMethod: "totp",
				Privilege: "root",
				Password:  "root",
				OptHeight: 1500,
				OptWidth:  1500,
				Hostname:  "127.0.3.1",
			}},
			wantErrMsg: "Service not created",
			wantStatus: false,
		},
		{
			name: "should pass",
			args: args{models.ConnectionParams{
				TotpCode:  testutils.GetTotpCode(testutils.MocktotpSEC),
				TfaMethod: "totp",
				Privilege: "root",
				Password:  "root",
				OptHeight: 1500,
				OptWidth:  1500,
				Hostname:  "127.0.0.1:2222",
			}},
			wantErrMsg: "",
			wantStatus: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErrMsg, gotStatus := connectSSHWS(t, &tt.args.params)
			if tt.wantStatus != gotStatus {
				t.Errorf("connectGuac() status = %t, wantErr %t", gotStatus, tt.wantStatus)
				return
			}

			if tt.wantErrMsg != gotErrMsg {
				t.Errorf("connectGuac() errMsg = %v, wantErr %v", gotErrMsg, tt.wantErrMsg)
				return
			}

		})
	}
}

func connectSSHWS(t *testing.T, params *models.ConnectionParams) (string, bool) {
	s := httptest.NewServer(testutils.AddTestUserContextWS(sshproxy.ConnectNewSSH))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	paramBytes, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}

	ws.WriteMessage(websocket.TextMessage, paramBytes)

	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	msgStr := string(msg)

	if !strings.Contains(string(msg), "Connecting...") {
		t.Fatalf(`expected "Connecting..."" string"`)
	}

	_, msg, err = ws.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	msgStr = string(msg)

	if strings.Contains(msgStr, `Service not created`) {
		return "Service not created", false
	}

	//TODO implement a generic format for detecting error and complete the test

	for i := 0; i < 50; i++ {
		_, _, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		err = ws.WriteMessage(websocket.TextMessage, []byte(`y`))
		if err != nil {
			t.Fatal(err)
		}
	}

	return "", true

}
