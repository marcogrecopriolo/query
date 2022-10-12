/*
Copyright 2021-Present Couchbase, Inc.

Use of this software is governed by the Business Source License included in
the file licenses/BSL-Couchbase.txt.  As of the Change Date specified in that
file, in accordance with the Business Source License, use of this software will
be governed by the Apache License, Version 2.0, included in the file
licenses/APL2.txt.
*/

package errors

const (
	E_OK                                      ErrorCode = 0
	E_SHELL_CONNECTION_REFUSED                ErrorCode = 100
	E_SHELL_UNSUPPORTED_PROTOCOL              ErrorCode = 101
	E_SHELL_NO_SUCH_HOST                      ErrorCode = 102
	E_SHELL_NO_HOST_IN_REQUEST_URL            ErrorCode = 103
	E_SHELL_UNKNOWN_PORT_TCP                  ErrorCode = 104
	E_SHELL_NO_ROUTE_TO_HOST                  ErrorCode = 105
	E_SHELL_UNREACHABLE_NETWORK               ErrorCode = 106
	E_SHELL_NO_CONNECTION                     ErrorCode = 107
	E_SHELL_DRIVER_OPEN                       ErrorCode = 108
	E_SHELL_INVALID_URL                       ErrorCode = 109
	E_SHELL_READ_FILE                         ErrorCode = 116
	E_SHELL_WRITE_FILE                        ErrorCode = 117
	E_SHELL_OPEN_FILE                         ErrorCode = 118
	E_SHELL_CLOSE_FILE                        ErrorCode = 119
	E_SHELL_INVALID_PASSWORD                  ErrorCode = 121
	E_SHELL_INVALID_USERNAME                  ErrorCode = 122
	E_SHELL_MISSING_CREDENTIAL                ErrorCode = 123
	E_SHELL_INVALID_CREDENTIAL                ErrorCode = 124
	E_SHELL_NO_SUCH_COMMAND                   ErrorCode = 136
	E_SHELL_NO_SUCH_PARAM                     ErrorCode = 137
	E_SHELL_TOO_MANY_ARGS                     ErrorCode = 138
	E_SHELL_TOO_FEW_ARGS                      ErrorCode = 139
	E_SHELL_STACK_EMPTY                       ErrorCode = 140
	E_SHELL_NO_SUCH_ALIAS                     ErrorCode = 141
	E_SHELL_BATCH_MODE                        ErrorCode = 142
	E_SHELL_STRING_WRITE                      ErrorCode = 143
	E_SHELL_OPERATION_TIMEOUT                 ErrorCode = 170
	E_SHELL_ROWS_SCAN                         ErrorCode = 171
	E_SHELL_JSON_MARSHAL                      ErrorCode = 172
	E_SHELL_JSON_UNMARSHAL                    ErrorCode = 173
	E_SHELL_DRIVER_QUERY_METHOD               ErrorCode = 174
	E_SHELL_WRITER_OUTPUT                     ErrorCode = 175
	E_SHELL_UNBALANCED_PAREN                  ErrorCode = 176
	E_SHELL_ROWS_CLOSE                        ErrorCode = 177
	E_SHELL_CMD_LINE_ARGS                     ErrorCode = 178
	E_SHELL_INVALID_INPUT_ARGUMENTS           ErrorCode = 179
	E_SHELL_ON_REFRESH                        ErrorCode = 180
	E_SHELL_INVALID_ARGUMENT                  ErrorCode = 181
	E_SHELL_INIT_FAILURE                      ErrorCode = 182
	E_SHELL_INVALID_PROTOCOL                  ErrorCode = 183
	E_SHELL_UNKNOWN                           ErrorCode = 199
	E_SERVICE_READONLY                        ErrorCode = 1000
	E_SERVICE_HTTP_UNSUPPORTED_METHOD         ErrorCode = 1010
	E_SERVICE_NOT_IMPLEMENTED                 ErrorCode = 1020
	E_SERVICE_UNRECOGNIZED_VALUE              ErrorCode = 1030
	E_SERVICE_BAD_VALUE                       ErrorCode = 1040
	E_SERVICE_MISSING_VALUE                   ErrorCode = 1050
	E_SERVICE_MULTIPLE_VALUES                 ErrorCode = 1060
	E_SERVICE_UNRECOGNIZED_PARAMETER          ErrorCode = 1065
	E_SERVICE_TYPE_MISMATCH                   ErrorCode = 1070
	E_SERVICE_TIMEOUT                         ErrorCode = 1080
	E_SERVICE_INVALID_JSON                    ErrorCode = 1100
	E_SERVICE_CLIENTID                        ErrorCode = 1110
	E_SERVICE_MEDIA_TYPE                      ErrorCode = 1120
	E_SERVICE_HTTP_REQ                        ErrorCode = 1130
	E_SERVICE_SCAN_VECTOR_BAD_LENGTH          ErrorCode = 1140
	E_SERVICE_SCAN_VECTOR_BAD_SEQUENCE_NUMBER ErrorCode = 1150
	E_SERVICE_SCAN_VECTOR_BADUUID             ErrorCode = 1155
	E_SERVICE_DECODE_NIL                      ErrorCode = 1160
	E_SERVICE_HTTP_METHOD                     ErrorCode = 1170
	E_SERVICE_SHUTTING_DOWN                   ErrorCode = 1180
	E_SERVICE_SHUT_DOWN                       ErrorCode = 1181
	E_SERVICE_USER_REQUEST_EXCEEDED           ErrorCode = 1191
	E_SERVICE_USER_REQUEST_RATE_EXCEEDED      ErrorCode = 1192
	E_SERVICE_USER_REQUEST_SIZE_EXCEEDED      ErrorCode = 1193
	E_SERVICE_USER_RESULT_SIZE_EXCEEDED       ErrorCode = 1194
	E_REQUEST_ERROR_LIMIT                     ErrorCode = 1195
	E_SERVICE_TENANT_THROTTLED                ErrorCode = 1196
	E_SERVICE_TENANT_MISSING                  ErrorCode = 1197
	E_SERVICE_TENANT_NOT_AUTHORIZED           ErrorCode = 1198
	E_SERVICE_TENANT_REJECTED                 ErrorCode = 1199
	E_ADMIN_CONNECTION                        ErrorCode = 2000
	E_ADMIN_START                             ErrorCode = 2001
	E_ADMIN_INVALIDURL                        ErrorCode = 2010
	E_ADMIN_DECODING                          ErrorCode = 2020
	E_ADMIN_ENCODING                          ErrorCode = 2030
	E_ADMIN_UNKNOWN_SETTING                   ErrorCode = 2031
	E_ADMIN_SETTING_TYPE                      ErrorCode = 2032
	E_ADMIN_GET_CLUSTER                       ErrorCode = 2040
	E_ADMIN_ADD_CLUSTER                       ErrorCode = 2050
	E_ADMIN_REMOVE_CLUSTER                    ErrorCode = 2060
	E_ADMIN_GET_NODE                          ErrorCode = 2070
	E_ADMIN_NO_NODE                           ErrorCode = 2080
	E_ADMIN_ADD_NODE                          ErrorCode = 2090
	E_ADMIN_REMOVE_NODE                       ErrorCode = 2100
	E_ADMIN_MAKE_METRIC                       ErrorCode = 2110
	E_ADMIN_AUTH                              ErrorCode = 2120
	E_ADMIN_ENDPOINT                          ErrorCode = 2130
	E_ADMIN_SSL_NOT_ENABLED                   ErrorCode = 2140
	E_ADMIN_CREDS                             ErrorCode = 2150
	E_COMPLETED_QUALIFIER_EXISTS              ErrorCode = 2160
	E_COMPLETED_QUALIFIER_UNKNOWN             ErrorCode = 2170
	E_COMPLETED_QUALIFIER_NOT_FOUND           ErrorCode = 2180
	E_COMPLETED_QUALIFIER_NOT_UNIQUE          ErrorCode = 2190
	E_COMPLETED_QUALIFIER_INVALID_ARGUMENT    ErrorCode = 2200
	E_ADMIN_BAD_SERVICE_PORT                  ErrorCode = 2210
	E_ADMIN_BODY                              ErrorCode = 2220
	E_PARSE_SYNTAX                            ErrorCode = 3000
	E_ERROR_CONTEXT                           ErrorCode = 3005
	E_PARSE_INVALID_ESCAPE_SEQUENCE           ErrorCode = 3006
	E_PARSE_INVALID_STRING                    ErrorCode = 3007
	E_PARSE_MISSING_CLOSING_QUOTE             ErrorCode = 3008
	E_PARSE_UNESCAPED_EMBEDDED_QUOTE          ErrorCode = 3009
	E_SEMANTICS                               ErrorCode = 3100
	E_JOIN_NEST_NO_JOIN_HINT                  ErrorCode = 3110
	E_JOIN_NEST_NO_USE_KEYS                   ErrorCode = 3120
	E_JOIN_NEST_NO_USE_INDEX                  ErrorCode = 3130
	_RETIRED_3140                                       = 3140
	E_MERGE_INSERT_NO_KEY                     ErrorCode = 3150
	E_MERGE_INSERT_MISSING_KEY                ErrorCode = 3160
	E_MERGE_MISSING_SOURCE                    ErrorCode = 3170
	E_MERGE_NO_INDEX_HINT                     ErrorCode = 3180
	E_MERGE_NO_JOIN_HINT                      ErrorCode = 3190
	E_MIXED_JOIN                              ErrorCode = 3200
	_RETIRED_3210                                       = 3210
	E_WINDOW_SEMANTIC                         ErrorCode = 3220
	E_ENTERPRISE_FEATURE                      ErrorCode = 3230
	_RETIRED_3240                                       = 3240
	E_ADVISE_UNSUPPORTED_STMT                 ErrorCode = 3250
	E_ADVISOR_PROJ_ONLY                       ErrorCode = 3255
	E_ADVISOR_NO_FROM                         ErrorCode = 3256
	E_MHDP_ONLY_FEATURE                       ErrorCode = 3260
	E_MISSING_USE_KEYS                        ErrorCode = 3261
	E_HAS_USE_INDEXES                         ErrorCode = 3262
	E_UPDATE_STAT_INVALID_INDEX_TYPE          ErrorCode = 3270
	E_UPDATE_STAT_INDEX_ALL_COLLECTION_ONLY   ErrorCode = 3271
	E_UPDATE_STAT_SELF_NOTALLOWED             ErrorCode = 3272
	E_CREATE_INDEX_NOT_INDEXABLE              ErrorCode = 3280
	E_CREATE_INDEX_ATTRIBUTE_MISSING          ErrorCode = 3281
	E_CREATE_INDEX_ATTRIBUTE                  ErrorCode = 3282
	E_FLATTEN_KEYS                            ErrorCode = 3283
	E_ALL_DISTINCT_NOT_ALLOWED                ErrorCode = 3284
	E_CREATE_INDEX_SELF_NOTALLOWED            ErrorCode = 3285
	E_JOIN_HINT_FIRST_FROM_TERM               ErrorCode = 3290
	E_PLAN                                    ErrorCode = 4000
	E_REPREPARE                               ErrorCode = 4001
	E_NO_TERM_NAME                            ErrorCode = 4010
	E_DUPLICATE_ALIAS                         ErrorCode = 4020
	E_UNKNOWN_FOR                             ErrorCode = 4025
	E_SUBQUERY_MISSING_KEYS                   ErrorCode = 4030
	E_SUBQUERY_MISSING_INDEX                  ErrorCode = 4035
	E_NO_SUCH_PREPARED                        ErrorCode = 4040
	E_NO_SUCH_PREPARED_WITH_CONTEXT           ErrorCode = 4041
	E_UNRECOGNIZED_PREPARED                   ErrorCode = 4050
	E_PREPARED_NAME                           ErrorCode = 4060
	E_PREPARED_DECODING                       ErrorCode = 4070
	E_PREPARED_ENCODING_MISMATCH              ErrorCode = 4080
	E_ENCODING_NAME_MISMATCH                  ErrorCode = 4090
	E_ENCODING_CONTEXT_MISMATCH               ErrorCode = 4091
	E_PREDEFINED_PREPARED_NAME                ErrorCode = 4092
	E_NO_INDEX_JOIN                           ErrorCode = 4100
	E_USE_KEYS_USE_INDEXES                    ErrorCode = 4110
	E_NO_PRIMARY_INDEX                        ErrorCode = 4120
	E_PRIMARY_INDEX_OFFLINE                   ErrorCode = 4125
	E_LIST_SUBQUERIES                         ErrorCode = 4130
	E_NOT_GROUP_KEY_OR_AGG                    ErrorCode = 4210
	E_INDEX_ALREADY_EXISTS                    ErrorCode = 4300
	E_AMBIGUOUS_META                          ErrorCode = 4310
	E_INDEXER_DESC_COLLATION                  ErrorCode = 4320
	E_PLAN_INTERNAL                           ErrorCode = 4321
	E_ALTER_INDEX                             ErrorCode = 4322
	E_NO_ANSI_JOIN                            ErrorCode = 4330
	E_PARTITION_INDEX_NOT_SUPPORTED           ErrorCode = 4340
	E_GSI                                     ErrorCode = 4350
	E_ENCODED_PLAN_NOT_ALLOWED                ErrorCode = 4400
	E_CBO                                     ErrorCode = 4600
	E_INDEX_STAT                              ErrorCode = 4610
	_RETIRED_4901                                       = 4901
	_RETIRED_4902                                       = 4902
	_RETIRED_4903                                       = 4903
	_RETIRED_4904                                       = 4904
	_RETIRED_4905                                       = 4905
	E_INTERNAL                                ErrorCode = 5000
	E_EXECUTION_PANIC                         ErrorCode = 5001
	E_EXECUTION_INTERNAL                      ErrorCode = 5002
	E_EXECUTION_PARAMETER                     ErrorCode = 5003
	E_PARSING                                 ErrorCode = 5004
	E_TEMP_FILE_QUOTA                         ErrorCode = 5005
	E_EXECUTION_KEY_VALIDATION                ErrorCode = 5006
	E_EVALUATION_ABORT                        ErrorCode = 5010
	E_EVALUATION                              ErrorCode = 5011
	E_EXPLAIN                                 ErrorCode = 5015
	E_GROUP_UPDATE                            ErrorCode = 5020
	E_INVALID_VALUE                           ErrorCode = 5030
	E_RANGE                                   ErrorCode = 5035
	W_DIVIDE_BY_ZERO                          ErrorCode = 5036
	E_DUPLICATE_FINAL_GROUP                   ErrorCode = 5040
	E_INSERT_KEY                              ErrorCode = 5050
	E_INSERT_VALUE                            ErrorCode = 5060
	E_INSERT_KEY_TYPE                         ErrorCode = 5070
	E_INSERT_OPTIONS_TYPE                     ErrorCode = 5071
	E_UPSERT_KEY                              ErrorCode = 5072
	E_UPSERT_KEY_ALREADY_MUTATED              ErrorCode = 5073
	E_UPSERT_VALUE                            ErrorCode = 5075
	E_UPSERT_KEY_TYPE                         ErrorCode = 5078
	E_UPSERT_OPTIONS_TYPE                     ErrorCode = 5079
	E_DELETE_ALIAS_MISSING                    ErrorCode = 5080
	E_DELETE_ALIAS_METADATA                   ErrorCode = 5090
	E_UPDATE_ALIAS_MISSING                    ErrorCode = 5100
	E_UPDATE_ALIAS_METADATA                   ErrorCode = 5110
	E_UPDATE_MISSING_CLONE                    ErrorCode = 5120
	E_UNNEST_INVALID_POSITION                 ErrorCode = 5180
	E_SCAN_VECTOR_TOO_MANY_SCANNED_BUCKETS    ErrorCode = 5190
	_RETIRED_5200                                       = 5200
	E_USER_NOT_FOUND                          ErrorCode = 5210
	E_ROLE_REQUIRES_KEYSPACE                  ErrorCode = 5220
	E_ROLE_TAKES_NO_KEYSPACE                  ErrorCode = 5230
	E_NO_SUCH_KEYSPACE                        ErrorCode = 5240
	E_NO_SUCH_SCOPE                           ErrorCode = 5241
	E_NO_SUCH_BUCKET                          ErrorCode = 5242
	E_ROLE_NOT_FOUND                          ErrorCode = 5250
	E_ROLE_ALREADY_PRESENT                    ErrorCode = 5260
	E_ROLE_NOT_PRESENT                        ErrorCode = 5270
	E_USER_WITH_NO_ROLES                      ErrorCode = 5280
	_RETIRED_5290                                       = 5290
	E_HASH_TABLE_PUT                          ErrorCode = 5300
	E_HASH_TABLE_GET                          ErrorCode = 5310
	E_MERGE_MULTI_UPDATE                      ErrorCode = 5320
	E_MERGE_MULTI_INSERT                      ErrorCode = 5330
	E_WINDOW_EVALUATION                       ErrorCode = 5340
	E_ADVISE_INDEX                            ErrorCode = 5350
	E_UPDATE_STATISTICS                       ErrorCode = 5360
	E_SUBQUERY_BUILD                          ErrorCode = 5370
	E_INDEX_LEADING_KEY_MISSING_NOT_SUPPORTED ErrorCode = 5380
	E_INDEX_NOT_IN_MEMORY                     ErrorCode = 5390
	E_MISSING_SYSTEMCBO_STATS                 ErrorCode = 5400
	E_INVALID_INDEX_NAME                      ErrorCode = 5410
	E_INDEX_NOT_FOUND                         ErrorCode = 5411
	E_INDEX_UPD_STATS                         ErrorCode = 5415
	E_TIME_PARSE                              ErrorCode = 5416
	E_JOIN_ON_PRIMARY_DOCS_EXCEEDED           ErrorCode = 5420
	E_MEMORY_QUOTA_EXCEEDED                   ErrorCode = 5500
	E_NIL_EVALUATE_PARAM                      ErrorCode = 5501
	E_BUCKET_ACTION                           ErrorCode = 5502
	W_MISSING_KEY                             ErrorCode = 5503
	E_NODE_QUOTA_EXCEEDED                     ErrorCode = 5600
	E_TENANT_QUOTA_EXCEEDED                   ErrorCode = 5601
	E_VALUE_RECONSTRUCT                       ErrorCode = 5700
	E_VALUE_INVALID                           ErrorCode = 5701
	E_VALUE_SPILL_CREATE                      ErrorCode = 5702
	E_VALUE_SPILL_READ                        ErrorCode = 5703
	E_VALUE_SPILL_WRITE                       ErrorCode = 5704
	E_VALUE_SPILL_SIZE                        ErrorCode = 5705
	E_VALUE_SPILL_SEEK                        ErrorCode = 5706
	E_SCHEDULER                               ErrorCode = 6001
	E_DUPLICATE_TASK                          ErrorCode = 6002
	E_TASK_RUNNING                            ErrorCode = 6003
	E_TASK_NOT_FOUND                          ErrorCode = 6004
	E_REWRITE                                 ErrorCode = 6500
	E_INFER_INVALID_OPTION                    ErrorCode = 7000
	E_INFER_OPTION_MUST_BE_NUMERIC            ErrorCode = 7001
	E_INFER_READING_NUMBER                    ErrorCode = 7002
	E_INFER_NO_KEYSPACE_DOCUMENTS             ErrorCode = 7003
	E_INFER_CREATE_RETRIEVER                  ErrorCode = 7004
	E_INFER_NO_RANDOM_ENTRY                   ErrorCode = 7005
	E_INFER_NO_RANDOM_DOCS                    ErrorCode = 7006
	E_INFER_MISSING_CONTEXT                   ErrorCode = 7007
	E_INFER_EXPRESSION_EVAL                   ErrorCode = 7008
	E_INFER_KEYSPACE_ERROR                    ErrorCode = 7009
	E_INFER_NO_SUITABLE_PRIMARY_INDEX         ErrorCode = 7010
	E_INFER_NO_SUITABLE_SECONDARY_INDEX       ErrorCode = 7011
	E_INFER_TIMEOUT                           ErrorCode = 7012
	E_INFER_SIZE_LIMIT                        ErrorCode = 7013
	E_INFER_NO_DOCUMENTS                      ErrorCode = 7014
	E_INFER_CONNECT                           ErrorCode = 7015
	E_INFER_GET_POOL                          ErrorCode = 7016
	E_INFER_GET_BUCKET                        ErrorCode = 7017
	E_INFER_INDEX_WARNING                     ErrorCode = 7018
	E_INFER_GET_RANDOM                        ErrorCode = 7019
	E_INFER_NO_RANDOM_SCAN                    ErrorCode = 7020
	E_DATASTORE_AUTHORIZATION                 ErrorCode = 10000
	E_FTS_MISSING_PORT_ERR                    ErrorCode = 10003
	E_NODE_INFO_ACCESS_ERR                    ErrorCode = 10004
	E_NODE_SERVICE_ERR                        ErrorCode = 10005
	E_FUNCTIONS_NOT_SUPPORTED                 ErrorCode = 10100
	E_MISSING_FUNCTION                        ErrorCode = 10101
	E_DUPLICATE_FUNCTION                      ErrorCode = 10102
	E_INTERNAL_FUNCTION                       ErrorCode = 10103
	E_ARGUMENTS_MISMATCH                      ErrorCode = 10104
	E_INVALID_FUNCTION_NAME                   ErrorCode = 10105
	E_FUNCTIONS_STORAGE                       ErrorCode = 10106
	E_FUNCTION_ENCODING                       ErrorCode = 10107
	E_FUNCTIONS_DISABLED                      ErrorCode = 10108
	E_FUNCTION_EXECUTION                      ErrorCode = 10109
	E_METAKV_CHANGE_COUNTER                   ErrorCode = 10110
	E_METAKV_INDEX                            ErrorCode = 10111
	E_TOO_MANY_NESTED_FUNCTIONS               ErrorCode = 10112
	E_INNER_FUNCTION_EXECUTION                ErrorCode = 10113
	E_LIBRARY_PATH_ERROR                      ErrorCode = 10114
	E_FUNCTION_LOADING                        ErrorCode = 10115
	E_EVALUATOR_LOADING                       ErrorCode = 10116
	E_EVALUATOR_INFLATING                     ErrorCode = 10117
	E_DATASTORE_INVALID_BUCKET_PARTS          ErrorCode = 10200
	E_QUERY_CONTEXT                           ErrorCode = 10201
	E_BUCKET_NO_DEFAULT_COLLECTION            ErrorCode = 10202
	E_NO_DATASTORE                            ErrorCode = 10203
	E_DATASTORE_INVALID_COLLECTION_PARTS      ErrorCode = 10204
	E_DATASTORE_INVALID_KEYSPACE_PARTS        ErrorCode = 10205
	E_DATASTORE_INVALID_PATH                  ErrorCode = 10206
	E_DATASTORE_INVALID_SCOPE_PARTS           ErrorCode = 10207
	E_BUCKET_UPDATER_MAX_ERRORS               ErrorCode = 10300
	E_BUCKET_UPDATER_NO_HEALTHY_NODES         ErrorCode = 10301
	E_BUCKET_UPDATER_STREAM_ERROR             ErrorCode = 10302
	E_BUCKET_UPDATER_AUTH_ERROR               ErrorCode = 10303
	E_BUCKET_UPDATER_CONNECTION_FAILED        ErrorCode = 10304
	E_BUCKET_UPDATER_ERROR_MAPPING            ErrorCode = 10305
	E_ADVISOR_SESSION_NOT_FOUND               ErrorCode = 10500
	E_ADVISOR_INVALID_ACTION                  ErrorCode = 10501
	E_ADVISOR_ACTION_MISSING                  ErrorCode = 10502
	E_ADVISOR_INVALID_ARGS                    ErrorCode = 10503
	E_SYSTEM_DATASTORE                        ErrorCode = 11000
	E_SYSTEM_KEYSPACE_NOT_FOUND               ErrorCode = 11002
	E_SYSTEM_NOT_IMPLEMENTED                  ErrorCode = 11003
	E_SYSTEM_NOT_SUPPORTED                    ErrorCode = 11004
	E_SYSTEM_IDX_NOT_FOUND                    ErrorCode = 11005
	E_SYSTEM_IDX_NO_DROP                      ErrorCode = 11006
	E_SYSTEM_STMT_NOT_FOUND                   ErrorCode = 11007
	E_SYSTEM_REMOTE_WARNING                   ErrorCode = 11008
	E_SYSTEM_UNABLE_TO_RETRIEVE               ErrorCode = 11009
	E_SYSTEM_UNABLE_TO_UPDATE                 ErrorCode = 11010
	E_SYSTEM_FILTERED_ROWS_WARNING            ErrorCode = 11011 // reused
	E_SYSTEM_MALFORMED_KEY                    ErrorCode = 11012
	E_SYSTEM_NO_BUCKETS                       ErrorCode = 11013
	E_INVALID_PREPARED_ADMIN_OP               ErrorCode = 11014
	E_CB_CONNECTION                           ErrorCode = 12000
	_RETIRED_12001                                      = 12001
	E_CB_NAMESPACE_NOT_FOUND                  ErrorCode = 12002
	E_CB_KEYSPACE_NOT_FOUND                   ErrorCode = 12003
	E_CB_PRIMARY_INDEX_NOT_FOUND              ErrorCode = 12004
	E_CB_INDEXER_NOT_IMPLEMENTED              ErrorCode = 12005
	E_CB_KEYSPACE_COUNT                       ErrorCode = 12006
	_RETIRED_12007                                      = 12007
	E_CB_BULK_GET                             ErrorCode = 12008
	E_CB_DML                                  ErrorCode = 12009
	_RETIRED_12010                                      = 12010
	E_CB_DELETE_FAILED                        ErrorCode = 12011
	E_CB_LOAD_INDEXES                         ErrorCode = 12012
	E_CB_BUCKET_TYPE_NOT_SUPPORTED            ErrorCode = 12013
	_RETIRED_12014                                      = 12014
	E_CB_INDEX_SCAN_TIMEOUT                   ErrorCode = 12015
	E_CB_INDEX_NOT_FOUND                      ErrorCode = 12016
	E_CB_GET_RANDOM_ENTRY                     ErrorCode = 12017
	E_UNABLE_TO_INIT_CB_AUTH                  ErrorCode = 12018
	E_AUDIT_STREAM_HANDLER_FAILED             ErrorCode = 12019
	E_CB_BUCKET_NOT_FOUND                     ErrorCode = 12020
	E_CB_SCOPE_NOT_FOUND                      ErrorCode = 12021
	E_CB_KEYSPACE_SIZE                        ErrorCode = 12022
	E_CB_SECURITY_CONFIG_NOT_PROVIDED         ErrorCode = 12023
	E_CB_CREATE_SYSTEM_BUCKET                 ErrorCode = 12024
	E_CB_BUCKET_CREATE_SCOPE                  ErrorCode = 12025
	E_CB_BUCKET_DROP_SCOPE                    ErrorCode = 12026
	E_CB_BUCKET_CREATE_COLLECTION             ErrorCode = 12027
	E_CB_BUCKET_DROP_COLLECTION               ErrorCode = 12028
	E_CB_BUCKET_FLUSH_COLLECTION              ErrorCode = 12029
	E_BINARY_DOCUMENT_MUTATION                ErrorCode = 12030
	E_DURABILITY_NOT_SUPPORTED                ErrorCode = 12031
	E_PRESERVE_EXPIRY_NOT_SUPPORTED           ErrorCode = 12032
	E_CAS_MISMATCH                            ErrorCode = 12033
	E_DML_MC                                  ErrorCode = 12034
	E_CB_NOT_PRIMARY_INDEX                    ErrorCode = 12035
	E_DML_INSERT                              ErrorCode = 12036
	E_ACCESS_DENIED                           ErrorCode = 12037
	E_WITH_INVALID_OPTION                     ErrorCode = 12038
	E_WITH_INVALID_TYPE                       ErrorCode = 12039
	_RETIRED_13010                                      = 13010
	_RETIRED_13011                                      = 13011
	E_DATASTORE_CLUSTER                       ErrorCode = 13012
	E_DATASTORE_UNABLE_TO_RETRIEVE_ROLES      ErrorCode = 13013
	E_DATASTORE_INSUFFICIENT_CREDENTIALS      ErrorCode = 13014
	E_DATASTORE_UNABLE_TO_RETRIEVE_BUCKETS    ErrorCode = 13015
	E_DATASTORE_NO_ADMIN                      ErrorCode = 13016
	E_INDEX_SCAN_SIZE                         ErrorCode = 14000
	E_FILE_DATASTORE                          ErrorCode = 15000
	E_FILE_NAMESPACE_NOT_FOUND                ErrorCode = 15001
	E_FILE_KEYSPACE_NOT_FOUND                 ErrorCode = 15002
	E_FILE_DUPLICATE_NAMESPACE                ErrorCode = 15003
	E_FILE_DUPLICATE_KEYSPACE                 ErrorCode = 15004
	E_FILE_NO_KEYS_INSERT                     ErrorCode = 15005
	E_FILE_KEY_EXISTS                         ErrorCode = 15006
	E_FILE_DML                                ErrorCode = 15007
	E_FILE_KEYSPACE_NOT_DIR                   ErrorCode = 15008
	E_FILE_IDX_NOT_FOUND                      ErrorCode = 15009
	E_FILE_NOT_SUPPORTED                      ErrorCode = 15010
	E_FILE_PRIMARY_IDX_NO_DROP                ErrorCode = 15011
	E_OTHER_DATASTORE                         ErrorCode = 16000
	E_OTHER_NAMESPACE_NOT_FOUND               ErrorCode = 16001
	E_OTHER_KEYSPACE_NOT_FOUND                ErrorCode = 16002
	E_OTHER_NOT_IMPLEMENTED                   ErrorCode = 16003
	E_OTHER_IDX_NOT_FOUND                     ErrorCode = 16004
	E_OTHER_IDX_NO_DROP                       ErrorCode = 16005
	E_OTHER_NOT_SUPPORTED                     ErrorCode = 16006
	E_OTHER_KEY_NOT_FOUND                     ErrorCode = 16007
	E_INFERENCER_NOT_FOUND                    ErrorCode = 16020
	E_OTHER_NO_BUCKETS                        ErrorCode = 16021
	E_SCOPES_NOT_SUPPORTED                    ErrorCode = 16022
	E_STAT_UPDATER_NOT_FOUND                  ErrorCode = 16030
	E_NO_FLUSH                                ErrorCode = 16040
	E_SS_IDX_NOT_FOUND                        ErrorCode = 16050
	E_SS_NOT_SUPPORTED                        ErrorCode = 16051
	E_SS_INACTIVE                             ErrorCode = 16052
	E_SS_INVALID                              ErrorCode = 16053
	E_SS_CONTINUE                             ErrorCode = 16054
	E_SS_CREATE                               ErrorCode = 16055
	E_SS_CANCEL                               ErrorCode = 16056
	E_SS_TIMEOUT                              ErrorCode = 16057
	E_SS_CID_GET                              ErrorCode = 16058
	E_SS_CONN                                 ErrorCode = 16059
	E_SS_FETCH_WAIT_TIMEOUT                   ErrorCode = 16060
	E_SS_WORKER_ABORT                         ErrorCode = 16061
	E_SS_FAILED                               ErrorCode = 16062
	E_SS_SPILL                                ErrorCode = 16063
	E_SS_VALIDATE                             ErrorCode = 16064
	E_TRAN_DATASTORE_NOT_SUPPORTED            ErrorCode = 17001
	E_TRAN_STATEMENT_NOT_SUPPORTED            ErrorCode = 17002
	E_TRAN_FUNCTION_NOT_SUPPORTED             ErrorCode = 17003
	E_TRANSACTION_CONTEXT                     ErrorCode = 17004
	E_TRAN_STATEMENT_OUT_OF_ORDER             ErrorCode = 17005
	E_START_TRANSACTION                       ErrorCode = 17006
	E_COMMIT_TRANSACTION                      ErrorCode = 17007
	E_ROLLBACK_TRANSACTION                    ErrorCode = 17008
	E_NO_SAVEPOINT                            ErrorCode = 17009
	E_TRANSACTION_EXPIRED                     ErrorCode = 17010
	E_TRANSACTION_RELEASED                    ErrorCode = 17011
	E_DUPLICATE_KEY                           ErrorCode = 17012
	E_TRANSACTION_INUSE                       ErrorCode = 17013
	E_KEY_NOT_FOUND                           ErrorCode = 17014
	E_SCAS_MISMATCH                           ErrorCode = 17015
	E_TRANSACTION_MEMORY_QUOTA_EXCEEDED       ErrorCode = 17016
	E_TRANSACTION_FETCH                       ErrorCode = 17017
	E_POST_COMMIT_TRANSACTION                 ErrorCode = 17018
	E_AMBIGUOUS_COMMIT_TRANSACTION            ErrorCode = 17019
	E_TRANSACTION_STAGING                     ErrorCode = 17020
	E_TRANSACTION_QUEUE_FULL                  ErrorCode = 17021
	E_POST_COMMIT_TRANSACTION_WARNING         ErrorCode = 17022
	E_TRANCE_NOTSUPPORTED                     ErrorCode = 17097
	E_MEMORY_ALLOCATION                       ErrorCode = 17098
	E_TRANSACTION                             ErrorCode = 17099
	E_DICT_INTERNAL                           ErrorCode = 18010
	E_INVALID_GSI_INDEXER                     ErrorCode = 18020
	E_INVALID_GSI_INDEX                       ErrorCode = 18030
	E_SYSTEM_COLLECTION                       ErrorCode = 18040
	E_DICTIONARY_ENCODING                     ErrorCode = 18050
	E_DICT_KEYSPACE_MISMATCH                  ErrorCode = 18060
	E_VIRTUAL_KS_NOT_SUPPORTED                ErrorCode = 19000
	E_VIRTUAL_KS_NOT_IMPLEMENTED              ErrorCode = 19001
	E_VIRTUAL_KS_IDXER_NOT_FOUND              ErrorCode = 19002
	E_VIRTUAL_IDX_NOT_FOUND                   ErrorCode = 19003
	E_VIRTUAL_IDXER_NOT_SUPPORTED             ErrorCode = 19004
	E_VIRTUAL_IDX_NOT_IMPLEMENTED             ErrorCode = 19005
	E_VIRTUAL_IDX_NOT_SUPPORTED               ErrorCode = 19006
	E_VIRTUAL_SCOPE_NOT_FOUND                 ErrorCode = 19007
	E_VIRTUAL_BUCKET_CREATE_SCOPE             ErrorCode = 19009
	E_VIRTUAL_BUCKET_DROP_SCOPE               ErrorCode = 19010
	E_VIRTUAL_KEYSPACE_NOT_FOUND              ErrorCode = 19011
	E_VIRTUAL_BUCKET_CREATE_COLLECTION        ErrorCode = 19012
	E_VIRTUAL_BUCKET_DROP_COLLECTION          ErrorCode = 19013
)
