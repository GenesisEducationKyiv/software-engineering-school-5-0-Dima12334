package handlers

import (
	"context"
	"ms-notification/internal/domain"
	"ms-notification/internal/service"

	pb "proto_stubs"
)

type NotificationGRPCHandler struct {
	pb.UnimplementedNotificationServiceServer
	emails service.Emails
}

func NewNotificationGRPCHandler(emails service.Emails) *NotificationGRPCHandler {
	return &NotificationGRPCHandler{emails: emails}
}

func (h *NotificationGRPCHandler) SendConfirmationEmail(
	ctx context.Context,
	req *pb.ConfirmationEmailRequest,
) (*pb.SendStatus, error) {
	input := domain.ConfirmationEmailInput{
		Email:            req.Email,
		ConfirmationLink: req.ConfirmationLink,
	}
	err := h.emails.SendConfirmationEmail(input)
	return &pb.SendStatus{Success: err == nil}, err
}

func (h *NotificationGRPCHandler) SendDailyForecastEmail(
	ctx context.Context,
	req *pb.DailyForecastEmailRequest,
) (*pb.SendStatus, error) {
	input := domain.WeatherForecastEmailInput[*domain.DayWeatherResponse]{
		Subscription:    fromProtoSubscription(req.Subscription),
		Weather:         fromProtoDayWeather(req.Weather),
		Date:            req.Date,
		UnsubscribeLink: req.UnsubscribeLink,
	}
	err := h.emails.SendWeatherForecastDailyEmail(input)
	return &pb.SendStatus{Success: err == nil}, err
}

func (h *NotificationGRPCHandler) SendHourlyForecastEmail(
	ctx context.Context,
	req *pb.HourlyForecastEmailRequest,
) (*pb.SendStatus, error) {
	input := domain.WeatherForecastEmailInput[*domain.WeatherResponse]{
		Subscription:    fromProtoSubscription(req.Subscription),
		Weather:         fromProtoWeather(req.Weather),
		Date:            req.Date,
		UnsubscribeLink: req.UnsubscribeLink,
	}
	err := h.emails.SendWeatherForecastHourlyEmail(input)
	return &pb.SendStatus{Success: err == nil}, err
}

func fromProtoSubscription(s *pb.Subscription) domain.Subscription {
	return domain.Subscription{
		Email: s.Email,
		City:  s.City,
	}
}

func fromProtoWeather(w *pb.Weather) *domain.WeatherResponse {
	return &domain.WeatherResponse{
		Temperature: w.Temperature,
		Humidity:    w.Humidity,
		Description: w.Description,
	}
}

func fromProtoDayWeather(d *pb.DayWeather) *domain.DayWeatherResponse {
	return &domain.DayWeatherResponse{
		SevenAM: *fromProtoWeather(d.SevenAm),
		TenAM:   *fromProtoWeather(d.TenAm),
		OnePM:   *fromProtoWeather(d.OnePm),
		FourPM:  *fromProtoWeather(d.FourPm),
		SevenPM: *fromProtoWeather(d.SevenPm),
		TenPM:   *fromProtoWeather(d.TenPm),
	}
}
