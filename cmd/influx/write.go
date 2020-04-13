package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	platform "github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/http"
	"github.com/influxdata/influxdb/v2/kit/signals"
	"github.com/influxdata/influxdb/v2/models"
	"github.com/influxdata/influxdb/v2/write"
	"github.com/spf13/cobra"
)

const (
	inputFormatCsv          = "csv"
	inputFormatLineProtocol = "lp"
)

type writeFlagsType struct {
	org       organization
	BucketID  string
	Bucket    string
	Precision string
	Format    string
	File      string
}

var writeFlags writeFlagsType

func cmdWrite(f *globalFlags, opt genericCLIOpts) *cobra.Command {
	cmd := opt.newCmd("write", fluxWriteF, true)
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Short = "Write points to InfluxDB"
	cmd.Long = `Write data to InfluxDB via stdin, or add an entire file specified with the -f flag`

	writeFlags.org.register(cmd, true)
	opts := flagOpts{
		{
			DestP:      &writeFlags.BucketID,
			Flag:       "bucket-id",
			Desc:       "The ID of destination bucket",
			Persistent: true,
		},
		{
			DestP:      &writeFlags.Bucket,
			Flag:       "bucket",
			Short:      'b',
			EnvVar:     "BUCKET_NAME",
			Desc:       "The name of destination bucket",
			Persistent: true,
		},
		{
			DestP:      &writeFlags.Precision,
			Flag:       "precision",
			Short:      'p',
			Default:    "ns",
			Desc:       "Precision of the timestamps of the lines",
			Persistent: true,
		},
	}
	opts.mustRegister(cmd)
	cmd.PersistentFlags().StringVar(&writeFlags.Format, "format", "", "Input format, either lp (Line Protocol) or csv (Comma Separated Values). Defaults to lp unless '.csv' extension")
	cmd.PersistentFlags().StringVarP(&writeFlags.File, "file", "f", "", "The path to the file to import")

	cmdDryRun := opt.newCmd("dryrun", fluxWriteDryrunF, false)
	cmdDryRun.Args = cobra.MaximumNArgs(1)
	cmdDryRun.Short = "Write to stdout instead of InfluxDB"
	cmdDryRun.Long = `Write protocol lines to stdout instead of InfluxDB. Troubleshoot conversion from CSV to line protocol.`
	cmd.AddCommand(cmdDryRun)
	return cmd
}

// createLineReader uses writeFlags and cli arguments to create a reader that produces line protocol
func (writeFlags *writeFlagsType) createLineReader(args []string) (r io.Reader, closer io.Closer, err error) {
	if len(args) > 0 && args[0][0] == '@' {
		// backward compatibility: @ in arg denotes a file
		writeFlags.File = args[0][1:]
	}

	if len(writeFlags.File) > 0 {
		f, err := os.Open(writeFlags.File)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open %q: %v", writeFlags.File, err)
		}
		closer = f
		r = f
		if len(writeFlags.Format) == 0 && strings.HasSuffix(writeFlags.File, ".csv") {
			writeFlags.Format = inputFormatCsv
		}
	} else if len(args) == 0 || args[0] == "-" {
		// backward compatibility: "-" also means stdin
		r = os.Stdin
	} else {
		r = strings.NewReader(args[0])
	}
	// validate input format
	if len(writeFlags.Format) > 0 && writeFlags.Format != inputFormatLineProtocol && writeFlags.Format != inputFormatCsv {
		return nil, nil, fmt.Errorf("unsupported input format: %s", writeFlags.Format)
	}

	if writeFlags.Format == inputFormatCsv {
		r = write.CsvToProtocolLines(r)
	}
	return r, closer, nil
}

func fluxWriteF(cmd *cobra.Command, args []string) error {
	// validate InfluxDB flags
	if err := writeFlags.org.validOrgFlags(&flags); err != nil {
		return err
	}

	if writeFlags.Bucket != "" && writeFlags.BucketID != "" {
		return fmt.Errorf("please specify one of bucket or bucket-id")
	}

	if !models.ValidPrecision(writeFlags.Precision) {
		return fmt.Errorf("invalid precision")
	}

	bs, err := newBucketService()
	if err != nil {
		return err
	}

	var filter platform.BucketFilter
	if writeFlags.BucketID != "" {
		filter.ID, err = platform.IDFromString(writeFlags.BucketID)
		if err != nil {
			return fmt.Errorf("failed to decode bucket-id: %v", err)
		}
	}
	if writeFlags.Bucket != "" {
		filter.Name = &writeFlags.Bucket
	}

	if writeFlags.org.id != "" {
		filter.OrganizationID, err = platform.IDFromString(writeFlags.org.id)
		if err != nil {
			return fmt.Errorf("failed to decode org-id id: %v", err)
		}
	}
	if writeFlags.org.name != "" {
		filter.Org = &writeFlags.org.name
	}

	ctx := context.Background()
	buckets, n, err := bs.FindBuckets(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to retrieve buckets: %v", err)
	}

	if n == 0 {
		if writeFlags.Bucket != "" {
			return fmt.Errorf("bucket %q was not found", writeFlags.Bucket)
		}

		if writeFlags.BucketID != "" {
			return fmt.Errorf("bucket with id %q does not exist", writeFlags.BucketID)
		}
	}
	bucketID, orgID := buckets[0].ID, buckets[0].OrgID

	// create line reader
	r, closer, err := writeFlags.createLineReader(args)
	if closer != nil {
		defer closer.Close()
	}
	if err != nil {
		return err
	}

	// write to InfluxDB
	s := write.Batcher{
		Service: &http.WriteService{
			Addr:               flags.Host,
			Token:              flags.Token,
			Precision:          writeFlags.Precision,
			InsecureSkipVerify: flags.skipVerify,
		},
	}
	ctx = signals.WithStandardSignals(ctx)
	if err := s.Write(ctx, orgID, bucketID, r); err != nil && err != context.Canceled {
		return fmt.Errorf("failed to write data: %v", err)
	}

	return nil
}

func fluxWriteDryrunF(cmd *cobra.Command, args []string) error {
	// create line reader
	r, closer, err := writeFlags.createLineReader(args)
	if closer != nil {
		defer closer.Close()
	}
	if err != nil {
		return err
	}
	// dry run
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}
	return nil
}
