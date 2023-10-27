package llmi18n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	// See https://github.com/jmorganca/ollama
	ollamaBaseURL          = "http://localhost:11434"
	ollamaEndpointGenerate = "/api/generate"
)

type prompt struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Stream  bool    `json:"stream"`
	Options options `json:"options"`
}

type options struct {
	// Enable Mirostat sampling for controlling perplexity. (default: 0, 0 = disabled, 1 = Mirostat, 2 = Mirostat 2.0)
	Mirostat int `json:"mirostat"`
	// Influences how quickly the algorithm responds to feedback from the generated text. A lower learning rate will result in slower adjustments, while a higher learning rate will make the algorithm more responsive. (Default: 0.1)
	MirostatEta float64 `json:"mirostat_eta"`
	// Controls the balance between coherence and diversity of the output. A lower value will result in more focused and coherent text. (Default: 5.0)
	MirostatTau float64 `json:"mirostat_tau"`
	// The temperature of the model. Increasing the temperature will make the model answer more creatively. (Default: 0.8)
	Temperature float64 `json:"temperature"`
	// Sets the random number seed to use for generation. Setting this to a specific number will make the model generate the same text for the same prompt. (Default: 0)
	Seed int `json:"seed"`
	// Sets the stop sequences to use.
	Stop []string `json:"stop"`
	// Reduces the probability of generating nonsense. A higher value (e.g. 100) will give more diverse answers, while a lower value (e.g. 10) will be more conservative. (Default: 40)
	TopK int `json:"top_k"`
	// Works together with top-k. A higher value (e.g., 0.95) will lead to more diverse text, while a lower value (e.g., 0.5) will generate more focused and conservative text. (Default: 0.9)
	TopP float64 `json:"top_p"`
}

// See https://github.com/jmorganca/ollama/blob/main/docs/api.md for additonal fields.
type response struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`

	// Final response only.
	TotalDuration time.Duration `json:"total_duration"`
	LoadDuration  time.Duration `json:"load_duration"`
}

func TranslatedQuotedStrings(s string) (string, error) {
	p := prompt{
		Model:  "mistral", // llama2, codellama, mistral
		Stream: false,
		Options: options{
			// Mirostat:    1,
			// MirostatEta: 0.1,
			// MirostatTau: 5.0,
			Temperature: 0.3,
			Seed:        42,
			// TopK:        3,
			// TopP:        0.1,
			// Stop:        []string{"Sure"},
		},
		Prompt: s + `

Translate the quoted strings above to de. Preserve the format as is. No introduction or conclusion is needed.

		
`,
	}

	var result string
	generate(p, func(r response) error {
		result = r.Response

		if r.Done {
			fmt.Println()
			fmt.Println()
			fmt.Printf("Total duration: %s\n", r.TotalDuration)
			fmt.Printf("Load duration: %s\n", r.LoadDuration)
		}
		return nil
	})

	return result, nil
}

func generate(p prompt, fn func(r response) error) error {
	url := ollamaBaseURL + ollamaEndpointGenerate

	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)

		return fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, buf.String())
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	for {
		var r response
		err := dec.Decode(&r)
		if err != nil {
			return err
		}
		err = fn(r)
		if err != nil {
			return err
		}
		if r.Done {
			break
		}
	}

	return nil
}
