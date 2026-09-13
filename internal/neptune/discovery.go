package neptune

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsneptune "github.com/aws/aws-sdk-go-v2/service/neptune"
)

func DiscoverWriterEndpoint(ctx context.Context, awsCfg aws.Config, clusterID string) (string, error) {
	client := awsneptune.NewFromConfig(awsCfg)
	input := &awsneptune.DescribeDBClustersInput{}
	if strings.TrimSpace(clusterID) != "" {
		input.DBClusterIdentifier = aws.String(strings.TrimSpace(clusterID))
	}

	paginator := awsneptune.NewDescribeDBClustersPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("describe Neptune clusters: %w", err)
		}

		for _, cluster := range page.DBClusters {
			endpoint := aws.ToString(cluster.Endpoint)
			if endpoint == "" {
				continue
			}
			if engine := strings.ToLower(aws.ToString(cluster.Engine)); engine != "" && !strings.Contains(engine, "neptune") {
				continue
			}
			if status := strings.ToLower(aws.ToString(cluster.Status)); status != "" && status != "available" {
				continue
			}

			port := int32(8182)
			if cluster.Port != nil && *cluster.Port > 0 {
				port = *cluster.Port
			}
			return neptuneHTTPSURL(endpoint, port), nil
		}
	}

	if clusterID != "" {
		return "", fmt.Errorf("no available Neptune writer endpoint found for cluster %q", clusterID)
	}
	return "", fmt.Errorf("no available Neptune writer endpoint found")
}

func neptuneHTTPSURL(endpoint string, port int32) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}

	parsed, err := url.Parse(endpoint)
	if err == nil && parsed.Scheme != "" {
		if parsed.Port() == "" {
			parsed.Host = parsed.Hostname() + ":" + strconv.Itoa(int(port))
		}
		return parsed.String()
	}

	return "https://" + endpoint + ":" + strconv.Itoa(int(port))
}
