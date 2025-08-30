package producers

import (
	"github.com/yoshapihoff/bricks/auth/internal/config"
	"github.com/yoshapihoff/bricks/auth/internal/kafka"
	sendEmail "github.com/yoshapihoff/bricks/auth/pkg/sendEmail.v1"
)

type ForgotPasswordEmailProducer struct {
	srProducer kafka.SRProducer
	topic      string
}

func NewForgotPasswordEmailProducer(cfg *config.Config) (*ForgotPasswordEmailProducer, error) {
	srProducer, err := kafka.NewProducer(cfg.Kafka.KafkaUrl, cfg.Kafka.SchemaRegistryUrl)
	if err != nil {
		return nil, err
	}
	return &ForgotPasswordEmailProducer{
		srProducer: srProducer,
		topic:      cfg.ForgotPasswordEmailSendingTopic,
	}, nil
}

func (p *ForgotPasswordEmailProducer) ProduceForgotPasswordEmail(email, resetPasswordToken string) (int64, error) {
	sendEmailMsg := &sendEmail.SendEmail{
		To:       []string{email},
		Subject:  "Forgot Password",
		Template: "forgot-password",
		Params:   map[string]string{"reset_password_token": resetPasswordToken},
	}
	return p.srProducer.ProduceMessage(sendEmailMsg, p.topic)
}

func (p *ForgotPasswordEmailProducer) Close() {
	p.srProducer.Close()
}
