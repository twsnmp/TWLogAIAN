package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"go.etcd.io/bbolt"
)

type AIConfigEnt struct {
	ID              string `json:"ID"`
	WeaviateURL     string `json:"WeaviateURL"`
	OllamaURL       string `json:"OllamaURL"`
	Text2vecModel   string `json:"Text2vecModel"`
	GenerativeModel string `json:"GenerativeModel"`
	ClassName       string `json:"ClassName"`
}

func (a *App) GetAIConfigs() []AIConfigEnt {
	ret := []AIConfigEnt{}
	a.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		if b == nil {
			return fmt.Errorf("ai bucket not found")
		}
		b.ForEach(func(k []byte, v []byte) error {
			var aiConfig AIConfigEnt
			err := json.Unmarshal(v, &aiConfig)
			if err == nil {
				ret = append(ret, aiConfig)
			} else {
				OutLog("ai config err=%v", err)
			}
			return nil
		})
		return nil
	})
	return ret
}

func (a *App) getAIConfigMap(weaviateURL string) map[string]bool {
	ret := make(map[string]bool)
	l := a.GetAIConfigs()
	for _, ac := range l {
		if ac.WeaviateURL == weaviateURL {
			key := fmt.Sprintf("%s\t%s\t%s", ac.OllamaURL, ac.GenerativeModel, ac.Text2vecModel)
			ret[key] = true
		}
	}
	return ret
}

func (a *App) SyncAIConfig(weaviateURL string) string {
	m := a.getAIConfigMap(weaviateURL)
	aiConfig := AIConfigEnt{
		ID:          fmt.Sprintf("%016x", time.Now().UnixNano()),
		WeaviateURL: weaviateURL,
	}
	client, err := getWeaviateClient(aiConfig)
	if err != nil {
		return err.Error()
	}
	schema, err := client.Schema().Getter().Do(context.Background())
	if err != nil {
		return err.Error()
	}
	for i, c := range schema.Classes {
		OutLog("sync ai config class=%s", c.Class)
		oa, err := jsonpath.Get(`$["generative-ollama"].apiEndpoint`, c.ModuleConfig)
		if err != nil {
			OutLog("sync ai config err=%v", err)
			continue
		}
		va, err := jsonpath.Get(`$["text2vec-ollama"].apiEndpoint`, c.ModuleConfig)
		if err != nil {
			OutLog("sync ai config err=%v", err)
			continue
		}
		if va != oa {
			OutLog("sync ai config url %v!=%v", va, oa)
			continue
		}
		o, ok := oa.(string)
		if !ok {
			OutLog("sync ai config err=%v", err)
			continue
		}
		gm, err := jsonpath.Get(`$["generative-ollama"].model`, c.ModuleConfig)
		if err != nil {
			OutLog("sync ai config err=%v", err)
			continue
		}
		vm, err := jsonpath.Get(`$["text2vec-ollama"].model`, c.ModuleConfig)
		if err != nil {
			OutLog("sync ai config err=%v", err)
			continue
		}
		g, ok := gm.(string)
		if !ok {
			OutLog("sync ai config err=%v", err)
			continue
		}
		v, ok := vm.(string)
		if !ok {
			OutLog("sync ai config err=%v", err)
			continue
		}
		key := fmt.Sprintf("%s\t%s\t%s", o, g, v)
		if _, ok := m[key]; ok {
			OutLog("sync ai config dup skip class=%s", c.Class)
			continue
		}
		aiConfig := AIConfigEnt{
			ID:              fmt.Sprintf("%016x", time.Now().UnixNano()+int64(i)),
			WeaviateURL:     weaviateURL,
			OllamaURL:       o,
			Text2vecModel:   v,
			GenerativeModel: g,
			ClassName:       c.Class,
		}
		j, err := json.Marshal(&aiConfig)
		if err != nil {
			OutLog("sync ai config err=%v", err)
			continue
		}
		err = a.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("ai"))
			if b == nil {
				return fmt.Errorf("ai bucket not found")
			}
			return b.Put([]byte(aiConfig.ID), j)
		})
		if err != nil {
			OutLog("sync ai config err=%v", err)
			continue
		}
	}
	return ""
}

func (a *App) AddAIConfig(aiConfig AIConfigEnt) string {
	aiConfig.ID = fmt.Sprintf("%016x", time.Now().UnixNano())
	j, err := json.Marshal(&aiConfig)
	if err != nil {
		return err.Error()
	}
	client, err := getWeaviateClient(aiConfig)
	if err != nil {
		return err.Error()
	}
	classObj := &models.Class{
		Class:      aiConfig.ClassName,
		Vectorizer: "text2vec-ollama",
		ModuleConfig: map[string]interface{}{
			"text2vec-ollama": map[string]interface{}{
				"apiEndpoint": aiConfig.OllamaURL,
				"model":       aiConfig.Text2vecModel,
			},
			"generative-ollama": map[string]interface{}{
				"apiEndpoint": aiConfig.OllamaURL,
				"model":       aiConfig.GenerativeModel,
			},
		},
	}
	err = client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
	if err != nil {
		return err.Error()
	}
	err = a.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		if b == nil {
			return fmt.Errorf("ai bucket not found")
		}
		return b.Put([]byte(aiConfig.ID), j)
	})
	if err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) TestAIConfigByID(id string) string {
	aiConfig := a.getAIConfig(id)
	return a.TestAIConfig(aiConfig)
}

func (a *App) TestAIConfig(aiConfig AIConfigEnt) string {
	client, err := getWeaviateClient(aiConfig)
	if err != nil {
		return err.Error()
	}
	ready, err := client.Misc().ReadyChecker().Do(context.Background())
	if err != nil {
		return err.Error()
	}
	if !ready {
		return "weaviate not ready"
	}
	return ""
}

func (a *App) DeleteAIConfig(id, title, message string) string {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"Yes(Local Only)", "Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil || result == "No" {
		return "No"
	}
	if result == "Yes" {
		aiConfig := a.getAIConfig(id)
		client, err := getWeaviateClient(aiConfig)
		if err == nil {
			err = client.Schema().ClassDeleter().WithClassName(aiConfig.ClassName).Do(context.Background())
		}
		if err != nil {
			return err.Error()
		}
	}
	err = a.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		if b == nil {
			return fmt.Errorf("ai bucket not found")
		}
		return b.Delete([]byte(id))
	})
	if err != nil {
		return err.Error()
	}
	return ""
}

type AIExportEnt struct {
	Category string   `json:"Category"`
	Descr    string   `json:"Descr"`
	Logs     []LogEnt `json:"Logs"`
}

func (a *App) ExportAILog(id string, data AIExportEnt) string {
	aiConfig := a.getAIConfig(id)
	client, err := getWeaviateClient(aiConfig)
	if err != nil {
		return err.Error()
	}
	objects := []*models.Object{}
	for _, l := range data.Logs {
		ts := time.Unix(0, l.Time)
		objects = append(objects, &models.Object{
			Class: aiConfig.ClassName,
			Properties: map[string]any{
				"timestamp": ts,
				"log":       l.All,
				"category":  data.Category,
				"descr":     data.Descr,
			},
		})
	}
	st := time.Now()
	batchRes, err := client.Batch().ObjectsBatcher().WithObjects(objects...).Do(context.Background())
	if err != nil {
		return err.Error()
	}
	for _, res := range batchRes {
		if res.Result.Errors != nil {
			OutLog("batch err=%v", res.Result.Errors.Error)
		}
	}
	OutLog("batch end len=%d dur=%v", len(objects), time.Since(st))
	return ""
}

type AIAnswer struct {
	Error  string `json:"Error"`
	Answer string `json:"Answer"`
}

func (a *App) AskAIAboutLog(id string, prompt, log string, limit int) AIAnswer {
	aiConfig := a.getAIConfig(id)
	client, err := getWeaviateClient(aiConfig)
	if err != nil {
		OutLog("ask ai err=%v", err)
		return AIAnswer{
			Error: err.Error(),
		}
	}
	ctx := context.Background()
	gs := graphql.NewGenerativeSearch().GroupedResult(prompt)
	concepts := strings.Fields(log)
	response, err := client.GraphQL().Get().
		WithClassName(aiConfig.ClassName).
		WithFields(
			graphql.Field{Name: "timestamp"},
			graphql.Field{Name: "log"},
			graphql.Field{Name: "category"},
			graphql.Field{Name: "descr"},
		).
		WithGenerativeSearch(gs).
		WithNearText(client.GraphQL().NearTextArgBuilder().
			WithConcepts(concepts)).
		WithLimit(limit).
		Do(ctx)
	if err != nil {
		OutLog("ask ai err=%v", err)
		return AIAnswer{
			Error: err.Error(),
		}
	}
	errA := []string{}
	for _, e := range response.Errors {
		errA = append(errA, fmt.Sprintf("%v", e))
	}
	if len(errA) > 0 {
		return AIAnswer{
			Error: strings.Join(errA, "<br>"),
		}
	}
	OutLog("ai response=%+v", response)
	r, err := jsonpath.Get("$..groupedResult", response.Data["Get"])
	if err != nil {
		return AIAnswer{
			Error: err.Error(),
		}
	}
	return AIAnswer{
		Answer: fmt.Sprintf("%v", r),
	}
}

func (a *App) getAIConfig(id string) AIConfigEnt {
	var ret AIConfigEnt
	a.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		if b == nil {
			return fmt.Errorf("ai bucket not found")
		}
		if v := b.Get([]byte(id)); v != nil {
			json.Unmarshal(v, &ret)
		}
		return nil
	})
	return ret
}

func getWeaviateClient(aiConfig AIConfigEnt) (*weaviate.Client, error) {
	u, err := url.Parse(aiConfig.WeaviateURL)
	if err != nil {
		return nil, err
	}
	cfg := weaviate.Config{
		Host:   u.Host,
		Scheme: u.Scheme,
	}
	return weaviate.NewClient(cfg)
}
