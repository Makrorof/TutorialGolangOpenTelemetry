package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

	Process1()

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

func Process1() {
	mine := func(ctx context.Context, wg *sync.WaitGroup, productID int) {
		defer wg.Done()

		tr := otel.Tracer("Miner")

		newCtx, span := tr.Start(ctx, "ProductID-"+fmt.Sprint(productID))
		defer span.End()

		//Her i farkli bir methodu tanimlar.
		for i := 0; i < 5; i++ { // 5 adet farkli islemi var diyelim
			func() { //Defer icin eklendi
				_, span := tr.Start(newCtx, "ProcessID:"+fmt.Sprint(i))
				defer span.End()

				time.Sleep(time.Duration(rand.Intn(10)) * time.Second) //Islem suresi

				if rand.Intn(50) > 25 { //Islem basarisiz.
					err := errors.New("random value 25 uzeri geldigi icin islem basarisiz sayildi.")
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return
				}

				// Islem basarili
				if i == 4 {
					priceList := []int{rand.Intn(55), rand.Intn(55), rand.Intn(55), rand.Intn(55), rand.Intn(55)}
					span.SetAttributes(attribute.IntSlice("miner.getPriceList", priceList))
				} else {
					span.SetAttributes(attribute.String("miner.ProcessID_"+fmt.Sprint(i)+".Count", fmt.Sprint(rand.Intn(55))))
				}
			}()
		}
	}

	for i := 0; i < 5; i++ {
		go func(index int) {
			for {
				tr := otel.Tracer("Miner")
				trContext, mainSpan := tr.Start(context.Background(), "Miner-ScanPackageID-"+fmt.Sprint(index))
				wg := sync.WaitGroup{}

				for i2 := 0; i2 < 10; i2++ {
					wg.Add(1)
					go mine(trContext, &wg, i2)
				}

				wg.Wait()
				mainSpan.End()
			}
		}(i)
	}
}
