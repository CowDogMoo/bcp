# bcp

`bcp` is a command-line tool that provides SCP-like functionality for
cloud systems using a blob store. It allows you to upload files to an
S3 bucket and download files from the bucket to a remote instance via
AWS Systems Manager (SSM).

## Usage

```shell
bcp source_directory ssm_instance_id:destination_directory
```

- source_directory: The local directory or file path that you want to upload
to the S3 bucket.
- ssm_instance_id: The ID of the remote instance where you want to download
the file.
- destination_directory: The directory on the remote instance where you want
to download the file.

## Prerequisites
- AWS CLI installed on your local machine.
- AWS CLI configured with appropriate access credentials.
- AWS CLI installed on the remote instance.

## Getting Started

1. Clone the repository:

    ```shell
    git clone $REPO_URL
    ```

1. Build the project:

    ```shell
    go build
    ```

1. Run the bcp command:

    ```shell
    ./bcp $SRC_DIR $SSM_INST_ID:$DEST_DIR
    ```

1. Verify that the file was uploaded to the S3 bucket:

    ```shell
    aws s3 ls s3://$BUCKET_NAME
    ```


