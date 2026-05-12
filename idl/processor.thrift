namespace go diffgram.processor

include "common.thrift"

enum MediaType {
    IMAGE = 1
    VIDEO = 2
    TEXT = 3
    AUDIO = 4
    SENSOR_FUSION = 5
    GEO_TIFF = 6
}

struct ProcessMediaRequest {
    1: required i64 project_id
    2: required i64 input_id
    3: required string media_type
    4: optional map<string, string> metadata
}

struct ProcessMediaResponse {
    1: required i64 input_id
    2: required string status
    3: optional string error_message
}

struct GetInputRequest {
    1: required i64 input_id
}

struct InputInfo {
    1: required i64 id
    2: required i64 project_id
    3: required string media_type
    4: required string status
    5: optional double percent_complete
    6: optional string original_filename
    7: optional i64 file_id
}

struct ListInputsRequest {
    1: required i64 project_id
    2: optional common.Pagination pagination
    3: optional string status_filter
}

struct ListInputsResponse {
    1: required list<InputInfo> inputs
    2: required common.PaginatedMeta meta
}

service ProcessorService {
    ProcessMediaResponse ProcessMedia(1: ProcessMediaRequest req)
    InputInfo GetInput(1: GetInputRequest req)
    ListInputsResponse ListInputs(1: ListInputsRequest req)
}
