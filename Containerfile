# Copyright 2025 John Casey
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build stage
FROM registry.fedoraproject.org/fedora-minimal:41 AS builder

# Install Go and build dependencies
RUN microdnf -y update && \
    microdnf -y install \
    golang \
    git \
    && microdnf clean all

# Create and set working directory
WORKDIR /app

# Set Go module environment variables
ENV GO111MODULE=on
ENV GOPROXY=direct
ENV GOSUMDB=off

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .
COPY cmd cmd

# Verify module structure and build
RUN ls -lR /app && \
    echo "=== Verifying module structure ===" && \
    go list -m && \
    go list ./... && \
    echo "=== Building application ===" && \ 
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o myshift ./cmd/myshift

# Runtime stage
FROM registry.fedoraproject.org/fedora-minimal:41

# Install ca-certificates for HTTPS calls to PagerDuty API
RUN microdnf -y update && \
    microdnf -y install ca-certificates && \
    microdnf clean all

# Create non-root user
RUN useradd -r -s /bin/false myshift

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/myshift .

# Change ownership to non-root user
RUN chown myshift:myshift /app/myshift
RUN mkdir -p /home/myshift/.config

# Switch to non-root user
USER myshift

# Set the entrypoint
ENTRYPOINT ["./myshift"]
#ENTRYPOINT ["/bin/bash"]

# Default command (can be overridden)
CMD ["repl"] 