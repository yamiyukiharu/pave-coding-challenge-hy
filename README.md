# Billing API

This project is a billing API built using [Encore](https://encore.dev) and [Temporal](https://temporal.io) to manage billing workflows. The API allows creating, updating, and querying bills and their line items.

## Features

1. Create new bills.
2. Add line items to an existing open bill.
3. Close an active bill and get total charged amount.
4. Reject adding line items if a bill is already closed.
5. Query open and closed bills by status and account ID.
6. Retrieve a bill along with all its line items.

## Prerequisites

To run this project locally, you’ll need to have the following installed:

- [Go](https://golang.org/doc/install)
- [Docker](https://www.docker.com/get-started)
- [Encore](https://encore.dev/docs/install)
- [Temporal CLI](https://docs.temporal.io/docs/server/quick-install)

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/billing-api.git
cd billing-api
