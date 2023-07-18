package server

import (
	"context"
	"fmt"
	"time"

	"github.com/arseniy96/bonus-program/internal/logger"
	"github.com/arseniy96/bonus-program/internal/services/accrual"
)

const TTL = 1 * time.Second

func (s *Server) OrdersWorker() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for {
		select {
		case orderWithTTL := <-s.OrdersQueue:
			// дёргаем сервис, который отвечает за обработку заказа
			// внутри сервиса отправляем запрос в accrual
			// смотрим на статус:
			// если статус не конечный, то возвращаем order как есть (???)
			// если статус конечный, то обновляем ордер и создаём запись в bonus_transactions
			if orderWithTTL.ttl.Before(time.Now()) {
				res, err := accrual.CheckOrder(s.Config.AccrualHost, orderWithTTL.order.OrderNumber)
				logger.Log.Info(fmt.Sprintf("%v", res))
				if err != nil {
					logger.Log.Errorf("accrual check order error: %v", err)
					s.OrdersQueue <- OrderWithTTL{
						order: orderWithTTL.order,
						ttl:   time.Now().Add(TTL),
					}
					continue
				}

				if res.Status == accrual.OrderStatusRegistered || res.Status == accrual.OrderStatusProcessing {
					// система ещё не обработала заказ – отправляем обратно в очередь
					logger.Log.Debugw("accrual has not processed the order yet",
						"order_number", orderWithTTL.order.OrderNumber,
						"current_accrual_status", res.Status)
					s.OrdersQueue <- OrderWithTTL{
						order: orderWithTTL.order,
						ttl:   time.Now().Add(TTL),
					}
					continue
				}
				err = s.repository.UpdateOrderStatus(ctx, orderWithTTL.order, res.Status, int(res.Accrual*100))
				if err != nil {
					logger.Log.Error(err)
					s.OrdersQueue <- OrderWithTTL{
						order: orderWithTTL.order,
						ttl:   time.Now().Add(TTL),
					}
					continue
				}
				continue
			}
			s.OrdersQueue <- orderWithTTL
			continue
		default:
			continue
		}
	}
}
