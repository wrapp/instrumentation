package lastseen

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type cloudWatchExporter struct {
	cloudwatch  *cloudwatch.CloudWatch
	serviceName string
	namespace   string
}

// CloudWatch creates a new last seen exporter and opens a AWS session
func CloudWatch(serviceName string, namespace string) ExporterFactory {
	return func() (Exporter, error) {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		// Create new cloudwatch client.
		cw := cloudwatch.New(sess)

		return &cloudWatchExporter{
			cloudwatch:  cw,
			serviceName: serviceName,
			namespace:   namespace,
		}, nil
	}
}

func (t *cloudWatchExporter) Export(ctx context.Context, field string, val LastSeen) error {
	_, err := t.cloudwatch.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String(t.namespace),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				Timestamp:  aws.Time(time.Unix(val.Value, 0)),
				MetricName: aws.String("LastSeen"),
				Unit:       aws.String(cloudwatch.StandardUnitNone),
				Value:      aws.Float64(1),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("ServiceName"),
						Value: aws.String(t.serviceName),
					},
					&cloudwatch.Dimension{
						Name:  aws.String("Action"),
						Value: aws.String(field),
					},
				},
			},
		},
	})
	return err
}
