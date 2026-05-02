namespace go diffgram.common

struct Pagination {
    1: required i32 page
    2: required i32 page_size
}

struct PaginatedMeta {
    1: required i64 total
    2: required i32 page
    3: required i32 page_size
}