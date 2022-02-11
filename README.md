# helm-s3-index-restore

The index.yaml of your helm chart bucket is empty? This tool helps you.


## Usage

### To install the binary
```
git clone https://github.com/dennisvandehoef/helm-s3-index-restore.git
cd helm-s3-index-restore
go build . && go install
```

### Run the tool

```
helm-s3-index-restore -bucket helmcharts-demo -directory my-demo -profile default
```

Arguments:
```
  -bucket string
        Required, the name of the S3 bucket. Example: for s3://helmcharts-demo/my-demo you enter helmcharts-demo
  -directory string
        Optional, the path to the directory in the S3 bucket that contains the index file that needs to be restored and the charts. In case this is not the root of the s3 bucket. Example: for s3://helmcharts-demo/my-demo you enter my-demo
  -profile string
        Required, the AWS profile configured locally to connect to AWS. Example: default
  -upload
        Upload your new index.yaml directly to S3
```
