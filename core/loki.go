package core

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func (loki *LogLoki) SendLogToLoki(logData LogTelemetry) error {
	lokiURL := fmt.Sprintf("%v:%v/loki/api/v1/push", loki.Url, loki.Port)
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

	jsonLog, err := json.Marshal(logData)
	if err != nil {
		return fmt.Errorf("erro ao converter log para JSON: %v", err.Error())
	}
	fmt.Println(string(jsonLog))

	params := HttpRequestParams{
		Method: POST,
		URL:    lokiURL,
		Headers: Headers{
			ContentType: "application/json",
		},
		Body: map[string]any{
			"streams": []map[string]any{
				{
					"stream": map[string]any{
						"app": loki.AppName,
					},
					"values": [][]string{
						{timestamp, string(jsonLog)},
					},
				},
			},
		},
	}
	res, err := SendHttpRequest(params)
	if err != nil {
		return fmt.Errorf("erro ao enviar log para o Loki: %v", err.Error())
	}
	defer res.Body.Close()

	log.Println("Log enviado para o Loki com sucesso!")
	return nil
}
