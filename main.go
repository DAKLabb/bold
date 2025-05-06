package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type SummaryInfo struct {
	Count int
	Total float64
}

type TransactionInfo struct {
	Documents []struct {
		ID            string    `json:"id"`
		CreatedOn     time.Time `json:"createdOn"`
		ModifiedOn    time.Time `json:"modifiedOn"`
		CustomerEmail string    `json:"customerEmail"`
		SalesOrderID  any       `json:"salesOrderId"`
		Voided        bool      `json:"voided"`
		TotalSales    struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalSales"`
		TotalNetSales struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalNetSales"`
		TotalNetShipping struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalNetShipping"`
		TotalTaxes struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalTaxes"`
		Total struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"total"`
		TotalNetPayment struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalNetPayment"`
		Payments []struct {
			ID     string `json:"id"`
			Amount struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"amount"`
			RefundedAmount struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"refundedAmount"`
			NetAmount struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"netAmount"`
			CreditCardType any    `json:"creditCardType"`
			Provider       string `json:"provider"`
			Refunds        []any  `json:"refunds"`
			ProcessingFees []struct {
				ID     string `json:"id"`
				Amount struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"amount"`
				AmountGatewayCurrency struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"amountGatewayCurrency"`
				ExchangeRate   int `json:"exchangeRate"`
				RefundedAmount struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"refundedAmount"`
				RefundedAmountGatewayCurrency struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"refundedAmountGatewayCurrency"`
				NetAmount struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"netAmount"`
				NetAmountGatewayCurrency struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"netAmountGatewayCurrency"`
				FeeRefunds []any `json:"feeRefunds"`
			} `json:"processingFees"`
			GiftCardID                    any       `json:"giftCardId"`
			PaidOn                        time.Time `json:"paidOn"`
			ExternalTransactionID         string    `json:"externalTransactionId"`
			ExternalTransactionProperties []any     `json:"externalTransactionProperties"`
			ExternalCustomerID            any       `json:"externalCustomerId"`
		} `json:"payments"`
		SalesLineItems      []any `json:"salesLineItems"`
		Discounts           []any `json:"discounts"`
		ShippingLineItems   []any `json:"shippingLineItems"`
		PaymentGatewayError any   `json:"paymentGatewayError"`
	} `json:"documents"`
	Pagination struct {
		NextPageURL    any     `json:"nextPageUrl"`
		NextPageCursor *string `json:"nextPageCursor"`
		HasNextPage    bool    `json:"hasNextPage"`
	} `json:"pagination"`
}

func main() {

	url := "https://api.squarespace.com/1.0/commerce/transactions"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	token := fmt.Sprintf("Bearer %s", os.Getenv("SS_API"))
	req.Header.Add("Authorization", token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var result TransactionInfo
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Can not unmarshal JSON. Err: %v", err)
	}

	if result.Pagination.HasNextPage {
		cursor := *result.Pagination.NextPageCursor
		for result.Pagination.HasNextPage {
			nextPage := fmt.Sprintf("url?cursor=%s", cursor)
			nextReq, err := http.NewRequest(method, nextPage, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
			nextReq.Header.Add("Authorization", token)

			res, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(body))

			var moreResults TransactionInfo
			if err := json.Unmarshal(body, &moreResults); err != nil {
				fmt.Println("Can not unmarshal JSON")
			}
			result.Documents = append(result.Documents, moreResults.Documents...)
			break
		}
	}

	var total float64
	for _, doc := range result.Documents {
		value, err := strconv.ParseFloat(doc.Payments[0].Amount.Value, 64)
		if err != nil {
			log.Fatal("Error parsing string:", err)
			return
		}
		total = total + value
	}

	summary := SummaryInfo{Count: len(result.Documents), Total: total}
	jsonData, err := json.MarshalIndent(summary, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write JSON data to file
	err = os.WriteFile("summary.json", jsonData, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("Done")
}
