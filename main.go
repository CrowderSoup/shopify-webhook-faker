package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type formData struct {
	Message    string
	Error      string
	Payload    string
	TargetURL  string
	Topic      string
	ShopDomain string
}

var pageTmpl = template.Must(template.New("page").Parse(`
{{define "page"}}
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Shopify Webhook Faker</title>
  <style>
    :root {
      --bg: #0d1117;
      --card: #161b22;
      --accent: #7bdcb5;
      --text: #e6edf3;
      --muted: #9ea9b8;
      --danger: #ff6b6b;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: "Space Grotesk", "Segoe UI", "Helvetica Neue", sans-serif;
      background: radial-gradient(circle at 20% 20%, rgba(123, 220, 181, 0.15), transparent 35%),
                  radial-gradient(circle at 80% 0%, rgba(123, 220, 181, 0.12), transparent 30%),
                  var(--bg);
      color: var(--text);
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 24px;
    }
    .card {
      width: min(900px, 100%);
      background: var(--card);
      border: 1px solid rgba(255, 255, 255, 0.06);
      border-radius: 18px;
      box-shadow: 0 15px 40px rgba(0, 0, 0, 0.45);
      padding: 28px;
    }
    h1 {
      margin: 0 0 10px;
      font-size: 24px;
      letter-spacing: -0.3px;
    }
    p {
      margin: 0 0 20px;
      color: var(--muted);
      line-height: 1.5;
    }
    form {
      display: grid;
      gap: 16px;
    }
    label {
      display: flex;
      flex-direction: column;
      gap: 6px;
      font-weight: 600;
      letter-spacing: 0.2px;
    }
    input, textarea {
      background: #0f141b;
      color: var(--text);
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 10px;
      padding: 12px 14px;
      font-size: 14px;
      font-family: "JetBrains Mono", "SFMono-Regular", Menlo, monospace;
    }
    textarea {
      min-height: 220px;
      resize: vertical;
    }
    .row {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
      gap: 12px;
    }
    .hint {
      color: var(--muted);
      font-weight: 400;
      font-size: 12px;
    }
    .button-row {
      display: flex;
      justify-content: flex-end;
      gap: 12px;
      align-items: center;
      flex-wrap: wrap;
    }
    button {
      background: linear-gradient(135deg, var(--accent), #5fb797);
      color: #0b0f14;
      border: none;
      border-radius: 12px;
      padding: 12px 18px;
      font-weight: 700;
      letter-spacing: 0.2px;
      cursor: pointer;
      transition: transform 0.08s ease, box-shadow 0.08s ease;
      box-shadow: 0 10px 24px rgba(123, 220, 181, 0.25);
    }
    button:hover { transform: translateY(-1px); }
    button:active { transform: translateY(0); }
    .alert {
      padding: 12px 14px;
      border-radius: 10px;
      font-weight: 600;
    }
    .alert.success {
      background: rgba(123, 220, 181, 0.12);
      border: 1px solid rgba(123, 220, 181, 0.35);
      color: var(--text);
    }
    .alert.error {
      background: rgba(255, 107, 107, 0.12);
      border: 1px solid rgba(255, 107, 107, 0.35);
      color: var(--danger);
    }
    @media (max-width: 600px) {
      .card { padding: 20px; }
    }
  </style>
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>
</head>
<body>
  <main class="card">
    <h1>Shopify Webhook Faker</h1>
    <p>Paste your app's shared secret, craft a JSON payload, and deliver a signed webhook to any URL. The request uses the <code>X-Shopify-Hmac-Sha256</code> header from Shopify's webhook docs.</p>

    <div id="flash">{{template "flash" .}}</div>

    <form
      method="POST"
      action="/send"
      hx-post="/send"
      hx-target="#flash"
      hx-swap="innerHTML"
    >
      <div class="row">
        <label>
          Shopify shared secret
          <input required name="secret" id="secret-input" type="password" placeholder="shpss_..." autocomplete="off">
          <span class="hint">Used only to sign the payload on this request.</span>
        </label>
        <label>
          Target URL
          <input required name="target" type="url" placeholder="https://your-app.example.com/webhooks" value="{{.TargetURL}}">
          <span class="hint">Endpoint that should receive the fake webhook.</span>
        </label>
      </div>

      <div class="row">
        <label>
          Shopify topic (optional)
          <input name="topic" type="text" placeholder="orders/create" value="{{.Topic}}">
          <span class="hint">Added to <code>X-Shopify-Topic</code>; defaults to orders/create.</span>
        </label>
        <label>
          Shop domain (optional)
          <input name="shopDomain" type="text" placeholder="example.myshopify.com" value="{{.ShopDomain}}">
          <span class="hint">Added to <code>X-Shopify-Shop-Domain</code>; defaults to example.myshopify.com.</span>
        </label>
      </div>

      <label>
        JSON body
        <textarea required name="payload" placeholder='{"example": "value"}'>{{.Payload}}</textarea>
      </label>

      <div class="button-row">
        <button type="submit">Send Signed Webhook</button>
      </div>
    </form>
  </main>
</body>
</html>
{{end}}

{{define "flash"}}
  {{if .Message}}<div class="alert success" role="status">{{.Message}}</div>{{end}}
  {{if .Error}}<div class="alert error" role="alert">{{.Error}}</div>{{end}}
{{end}}
`))

func main() {
	log.Println("Starting server on http://localhost:8080 ...")
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/send", sendHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := formData{
		Topic:      "orders/create",
		ShopDomain: "example.myshopify.com",
	}
	render(w, data)
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	hxRequest := r.Header.Get("HX-Request") == "true"

	if err := r.ParseForm(); err != nil {
		respond(w, hxRequest, http.StatusBadRequest, formData{Error: "Failed to parse form: " + err.Error()})
		return
	}

	secret := strings.TrimSpace(r.FormValue("secret"))
	target := strings.TrimSpace(r.FormValue("target"))
	payload := strings.TrimSpace(r.FormValue("payload"))
	topic := strings.TrimSpace(r.FormValue("topic"))
	shopDomain := strings.TrimSpace(r.FormValue("shopDomain"))

	data := formData{
		Payload:    payload,
		TargetURL:  target,
		Topic:      defaultValue(topic, "orders/create"),
		ShopDomain: defaultValue(shopDomain, "example.myshopify.com"),
	}

	switch {
	case secret == "":
		data.Error = "Secret is required."
		respond(w, hxRequest, http.StatusBadRequest, data)
		return
	case target == "":
		data.Error = "Target URL is required."
		respond(w, hxRequest, http.StatusBadRequest, data)
		return
	case payload == "":
		data.Error = "Payload is required."
		respond(w, hxRequest, http.StatusBadRequest, data)
		return
	}

	if _, err := url.ParseRequestURI(target); err != nil {
		data.Error = "Target URL must be valid: " + err.Error()
		respond(w, hxRequest, http.StatusBadRequest, data)
		return
	}

	if !json.Valid([]byte(payload)) {
		data.Error = "Payload must be valid JSON."
		respond(w, hxRequest, http.StatusBadRequest, data)
		return
	}

	signature := signPayload(secret, []byte(payload))

	req, err := http.NewRequest(http.MethodPost, target, strings.NewReader(payload))
	if err != nil {
		data.Error = "Failed to create request: " + err.Error()
		respond(w, hxRequest, http.StatusBadRequest, data)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Hmac-Sha256", signature)
	req.Header.Set("X-Shopify-Topic", data.Topic)
	req.Header.Set("X-Shopify-Shop-Domain", data.ShopDomain)
	req.Header.Set("User-Agent", "Shopify-Webhook-Faker/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		data.Error = "Request failed: " + err.Error()
		respond(w, hxRequest, http.StatusBadGateway, data)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		data.Error = fmt.Sprintf("Received status %d but failed to read response: %v", resp.StatusCode, err)
		respond(w, hxRequest, http.StatusBadGateway, data)
		return
	}

	data.Message = fmt.Sprintf("Webhook sent. Status: %s. Response: %s", resp.Status, strings.TrimSpace(string(respBody)))
	respond(w, hxRequest, http.StatusOK, data)
}

func signPayload(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func defaultValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func render(w http.ResponseWriter, data formData) {
	setHTMLContentType(w)
	if err := pageTmpl.ExecuteTemplate(w, "page", data); err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

func renderFlash(w http.ResponseWriter, data formData) {
	setHTMLContentType(w)
	if err := pageTmpl.ExecuteTemplate(w, "flash", data); err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

func respond(w http.ResponseWriter, hxRequest bool, status int, data formData) {
	setHTMLContentType(w)
	if status > 0 {
		w.WriteHeader(status)
	}

	if hxRequest {
		renderFlash(w, data)
		return
	}

	render(w, data)
}

func setHTMLContentType(w http.ResponseWriter) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
}
