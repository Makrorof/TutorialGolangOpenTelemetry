package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	//rand.Seed(time.Hour.Nanoseconds())

	//Create jaeger Shutdown system
	{
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := Tracer.Shutdown(ctx); err != nil {
				log.Fatal(err)
			}
		}(ctx)
	}

	Process()

	for {
		time.Sleep(time.Second * 5)

	}
}

func Process() {
	for i := 0; i < 5; i++ {
		go func(index int) {
			tr := otel.Tracer("Process-Handler-ID-" + fmt.Sprint(index))
			//Span altinda span birikmesini istiyorsan bir span'dan uretilmis context verisini diger spanlari uretirken ver. Zaten infosunda yaziyor.
			trContext, _ := tr.Start(context.Background(), "Process-Handler-ID-"+fmt.Sprint(index))

			trIndex := 0
			for {
				rnd := rand.Intn(5)

				time.Sleep(time.Duration(rnd) * time.Second)

				fmt.Println("Process..")

				func() {
					for i2 := 0; i2 < rand.Intn(15); i2++ {
						_, span := tr.Start(trContext, "Process-ID-"+fmt.Sprint(trIndex)+"-SubID-"+fmt.Sprint(i2))
						span.SetAttributes(attribute.Key("first").String("firstValue"))
						span.SetAttributes(attribute.Key("Index:" + fmt.Sprint(i2)).String("Value: " + fmt.Sprint(i2+1)))

						time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
						span.End()

						log.Println("New Span")
					}
				}()

				trIndex++
				if trIndex == 1 {
					trIndex = 0
					trContext, _ = tr.Start(context.Background(), "Process-Handler-ID-"+fmt.Sprint(index))
					tr = otel.Tracer("Process-Handler-ID-" + fmt.Sprint(index))
				}
			}
		}(i)
	}
}
