# Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
FROM microsoft/windowsservercore:latest

MAINTAINER Amazon Web Services, Inc.

SHELL ["powershell", "-command"]

HEALTHCHECK --interval=1s --timeout=1s --retries=3 CMD echo hello

ENTRYPOINT ["powershell","-command"]
CMD ["Start-Sleep -s 2000"]
