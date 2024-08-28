package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"waitq/api/internal/server/http"
	"waitq/api/internal/service/domain"
	"waitq/api/internal/storage/postgres"
	"waitq/api/pkg/stripe"
)

type Options struct {
	Port                           int    `help:"Port to listen on" short:"p" default:"8080"`
	DatabaseURL                    string `help:"Database URL" short:"d"`
	SupabaseHost                   string `help:"Supabase Host" short:"s"`
	SupabaseServiceKey             string `help:"Supabase Service Key" short:"k"`
	ApiName                        string `help:"API Name" short:"n"`
	ApiVersion                     string `help:"API Version" short:"v"`
	ResendAPIKey                   string `help:"Resend API Key" short:"r"`
	StripeAPIKey                   string `help:"Stripe API Key" short:"S"`
	StripeSubscriptionCallbackPath string `help:"Stripe Subscription Callback Path" short:"C"`
	BaseAPIURL                     string `help:"Base API URL" short:"B"`
	StripeWebhookSecret            string `help:"Stripe Webhook Secret" short:"W"`
}

func (o *Options) config() {
	if port, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		o.Port = port
	}

	o.DatabaseURL = os.Getenv("DATABASE_URL")
	o.SupabaseHost = os.Getenv("SUPABASE_HOST")
	o.SupabaseServiceKey = os.Getenv("SUPABASE_SERVICE_KEY")
	o.ApiName = os.Getenv("API_NAME")
	o.ApiVersion = os.Getenv("API_VERSION")
	o.ResendAPIKey = os.Getenv("RESEND_API_KEY")
	o.StripeAPIKey = os.Getenv("STRIPE_API_KEY")
	o.BaseAPIURL = os.Getenv("BASE_API_URL")
	o.StripeSubscriptionCallbackPath = os.Getenv("STRIPE_SUBSCRIPTION_CALLBACK_PATH")
	o.StripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
}

func main() {
	// Load environment variables from .env.local
	err := godotenv.Load(".env.local")
	if err != nil {
		fmt.Println("Error loading .env.local file")
	}

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		options.config()

		ctx := context.Background()
		logger := zap.New(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionConfig().EncoderConfig),
				zapcore.AddSync(os.Stdout), zap.InfoLevel))

		stripeClient := stripe.NewClient(options.StripeAPIKey)

		postgresConfig := postgres.NewConfig(options.DatabaseURL)
		repositories := postgres.NewRepository(postgresConfig, ctx, logger)
		services := domain.NewService(repositories, stripeClient)

		supabaseClient, err := supabase.NewClient(
			options.SupabaseHost,
			options.SupabaseServiceKey,
			&supabase.ClientOptions{},
		)
		if err != nil {
			logger.Fatal("Failed to create supabase client", zap.Error(err))
		}

		server := http.NewServer(
			services,
			options.ApiName,
			options.ApiVersion,
			logger,
			supabaseClient,
			options.StripeWebhookSecret,
		)

		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			server.Serve(fmt.Sprintf(":%d", options.Port))
		})
	})

	cli.Run()
}
