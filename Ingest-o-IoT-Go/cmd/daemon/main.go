// Comando daemon: simulador de sensores IIoT que publica telemetria via MQTT.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Reading é o payload publicado por cada sensor.
type Reading struct {
	SensorID string  `json:"sensor_id"`
	Time     string  `json:"time"` // RFC3339Nano, UTC
	Value    float64 `json:"value"`
}

func main() {
	broker := flag.String("broker", "tcp://localhost:1883", "endereço do broker MQTT")
	sensors := flag.Int("sensors", 10, "número de sensores simulados")
	interval := flag.Duration("interval", time.Second, "intervalo de amostragem por sensor")
	qos := flag.Int("qos", 0, "QoS do MQTT (0, 1 ou 2)")
	prefix := flag.String("topic-prefix", "iiot/sensors", "prefixo do tópico")
	flag.Parse()

	if *qos < 0 || *qos > 2 || *sensors < 1 || *interval <= 0 {
		log.Fatal("parâmetros inválidos: qos 0-2, sensors >= 1, interval > 0")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	opts := mqtt.NewClientOptions().
		AddBroker(*broker).
		SetClientID(fmt.Sprintf("daemon-%d", os.Getpid())).
		SetAutoReconnect(true).
		SetConnectRetry(true)
	client := mqtt.NewClient(opts)
	if t := client.Connect(); t.Wait() && t.Error() != nil {
		log.Fatalf("falha ao conectar no broker: %v", t.Error())
	}
	defer client.Disconnect(250)

	var sent, failed atomic.Uint64
	var wg sync.WaitGroup
	for i := 1; i <= *sensors; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runSensor(ctx, client, fmt.Sprintf("sensor-%03d", id), *prefix, byte(*qos), *interval, &sent, &failed)
		}(i)
	}

	log.Printf("%d sensores publicando em %s/<id>/telemetry a cada %s (qos=%d)", *sensors, *prefix, *interval, *qos)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			log.Printf("encerrado: %d mensagens enviadas, %d falhas", sent.Load(), failed.Load())
			return
		case <-ticker.C:
			log.Printf("enviadas=%d falhas=%d", sent.Load(), failed.Load())
		}
	}
}

// runSensor simula uma grandeza física como um passeio aleatório em torno de uma base.
func runSensor(ctx context.Context, c mqtt.Client, id, prefix string, qos byte, every time.Duration, sent, failed *atomic.Uint64) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	value := 20 + rng.Float64()*10
	topic := fmt.Sprintf("%s/%s/telemetry", prefix, id)

	// Defasagem inicial aleatória evita que todos os sensores publiquem no mesmo instante.
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Duration(rng.Int63n(int64(every)))):
	}

	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		value += rng.NormFloat64() * 0.2
		payload, _ := json.Marshal(Reading{
			SensorID: id,
			Time:     time.Now().UTC().Format(time.RFC3339Nano),
			Value:    value,
		})
		t := c.Publish(topic, qos, false, payload)
		if qos == 0 {
			sent.Add(1) // QoS 0 não tem confirmação
		} else if t.WaitTimeout(2*time.Second) && t.Error() == nil {
			sent.Add(1)
		} else {
			failed.Add(1)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
