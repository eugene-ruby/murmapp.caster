package app

import (
	"context"
	"database/sql"
	"log"

	"murmapp.caster/internal/config"
	"murmapp.caster/internal/rabbitmqinit"
	"murmapp.caster/internal/registration"
	"murmapp.caster/internal/storewriter"
	"murmapp.caster/internal/telegramout"

	"github.com/eugene-ruby/xconnect/rabbitmq"
	"github.com/eugene-ruby/xconnect/redisstore"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/streadway/amqp"
)

type App struct {
	rabbitmq *AppRabbitMQ
	redis    *AppRedis
	workers  *Worker
}

type AppRabbitMQ struct {
	conn    *amqp.Connection
	rawCh   *amqp.Channel
	channel rabbitmq.Channel
}

type WorkerDeps struct {
	Channel rabbitmq.Channel
	Config  *config.Config
	Store   *redisstore.Store
	DB      *sql.DB
}

func (a *AppRabbitMQ) Close() {
	if a.rawCh != nil {
		_ = a.rawCh.Close()
	}
	if a.conn != nil {
		_ = a.conn.Close()
	}
}

type AppRedis struct {
	store *redisstore.Store
	rdb   *redis.Client
}

func (r *AppRedis) Close() {
	if r.rdb != nil {
		_ = r.rdb.Close()
	}
}

type Worker struct {
	workerRegistration *rabbitmq.Worker
	workerTelegram     *rabbitmq.Worker
	workerStoreWriter  *rabbitmq.Worker
}

func (w *Worker) Start(ctx context.Context) error {
	if err := w.workerRegistration.Start(ctx); err != nil {
		return err
	}
	if err := w.workerTelegram.Start(ctx); err != nil {
		return err
	}
	if err := w.workerStoreWriter.Start(ctx); err != nil {
		return err
	}
	return nil
}

func (w *Worker) Wait() {
	w.workerRegistration.Wait()
	w.workerTelegram.Wait()
	w.workerStoreWriter.Wait()
}

func New(cfg config.Config) (*App, error) {
	pgDB, err := sql.Open("pgx", cfg.PostgreSQL.DSN)
	if err != nil {
		log.Fatalf("❌ can't connect to PostgreSQL: %v", err)
	}

	rmq, ch, err := initRabbitMQ(cfg)
	if err != nil {
		return nil, err
	}

	if err := rabbitmqinit.DeclareExchanges(ch); err != nil {
		return nil, err
	}

	appRedis := initRedis(cfg)

	workDeps := &WorkerDeps{
		DB:      pgDB,
		Channel: ch,
		Config:  &cfg,
		Store:   appRedis.store,
	}

	regWorker := newWorkerRegistration(workDeps)
	tgWorker := newWorkerTelegram(workDeps)
	storeWriter := newWorkerStoreWriter(workDeps)

	return &App{
		rabbitmq: rmq,
		redis:    appRedis,
		workers: &Worker{
			workerRegistration: regWorker,
			workerTelegram:     tgWorker,
			workerStoreWriter:  storeWriter,
		},
	}, nil
}

func (a *App) StartWorker(ctx context.Context) error {
	return a.workers.Start(ctx)
}

func (a *App) Wait() {
	a.workers.Wait()
}

func (a *App) Shutdown() {
	a.rabbitmq.Close()
	a.redis.Close()
}

func initRabbitMQ(cfg config.Config) (*AppRabbitMQ, rabbitmq.Channel, error) {
	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	wrapped := rabbitmq.WrapAMQPChannel(ch)
	return &AppRabbitMQ{
		conn:    conn,
		rawCh:   ch,
		channel: wrapped,
	}, wrapped, nil
}

func initRedis(cfg config.Config) *AppRedis {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.URL,
	})
	adapter := redisstore.NewGoRedisAdapter(rdb)
	store := redisstore.New(adapter)
	return &AppRedis{rdb: rdb, store: store}
}

func newWorkerRegistration(deps *WorkerDeps) *rabbitmq.Worker {
	outboundHandler := &registration.OutboundHandler{
		Config:  deps.Config,
		Channel: deps.Channel,
	}
	return rabbitmq.NewWorker(deps.Channel, rabbitmq.WorkerConfig{
		Queue:          "murmapp.caster.webhook.registration",
		ConsumerTag:    "caster_registration",
		AutoAck:        true,
		Declare:        true,
		BindRoutingKey: "webhook.registration",
		BindExchange:   "murmapp",
		Handler: func(d rabbitmq.Delivery) error {
			log.Printf("[worker] registration: %d bytes", len(d.Body))
			registration.HandleRegistration(d.Body, outboundHandler)
			return nil
		},
	})
}

func newWorkerTelegram(deps *WorkerDeps) *rabbitmq.Worker {
	outboundHandler := &telegramout.OutboundHandler{
		Config: deps.Config,
		Store:  deps.Store,
		DB:     deps.DB,
	}

	return rabbitmq.NewWorker(deps.Channel, rabbitmq.WorkerConfig{
		Queue:          "murmapp.caster.telegram.messages.out",
		ConsumerTag:    "caster_telegram_out",
		AutoAck:        true,
		Declare:        true,
		BindRoutingKey: "telegram.messages.out",
		BindExchange:   "murmapp",
		Handler: func(d rabbitmq.Delivery) error {
			log.Printf("[worker] telegram out: %d bytes", len(d.Body))
			telegramout.HandleMessageOut(d.Body, outboundHandler)
			return nil
		},
	})
}

func newWorkerStoreWriter(deps *WorkerDeps) *rabbitmq.Worker {
	handler := &storewriter.Handler{
		DB:                      deps.DB,
		TelegramIdEncryptionKey: deps.Config.Encryption.TelegramIdEncryptionKey,
		PrivateKey:              deps.Config.Encryption.PrivateRSAEncryptionKey,
	}
	return rabbitmq.NewWorker(deps.Channel, rabbitmq.WorkerConfig{
		Queue:          "murmapp.caster.telegram.encrypted.id",
		ConsumerTag:    "caster_storewriter",
		AutoAck:        true,
		Declare:        true,
		BindRoutingKey: "telegram.encrypted.id",
		BindExchange:   "murmapp",
		Handler: func(d rabbitmq.Delivery) error {
			log.Printf("[worker] storewriter: %d bytes", len(d.Body))
			storewriter.HandleEncryptedID(d.Body, handler)
			return nil
		},
	})
}
