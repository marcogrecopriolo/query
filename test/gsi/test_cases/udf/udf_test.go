//  Copyright 2023-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package udf

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/couchbase/query/test/gsi"
)

func TestUDFs(t *testing.T) {
	if strings.ToLower(os.Getenv("GSI_TEST")) != "true" {
		return
	}

	qc := start_cs()

	fmt.Println("\n\nInserting values into Bucket \n\n ")
	runMatch("insert.json", false, false, qc, t)

	runStmt(qc, "CREATE PRIMARY INDEX ON customer")

	runMatch("case_inline_udf_tests.json", false, true, qc, t)

	runMatch("case_inline_udf_bugs.json", false, true, qc, t)

	// Drop functions created in the inline UDF tests
	runStmt(qc, "DROP FUNCTION inline1 IF EXISTS")
	runStmt(qc, "DROP FUNCTION inline2 IF EXISTS")
	runStmt(qc, "DROP FUNCTION inline3 IF EXISTS")

	runMatch("case_n1ql_managed_js_udf_tests.json", false, true, qc, t)

	// Drop functions created in the N1QL managed JS UDF tests
	runStmt(qc, "DROP FUNCTION n1qlJS1 IF EXISTS")
	runStmt(qc, "DROP FUNCTION n1qlJS2 IF EXISTS")
	runStmt(qc, "DROP FUNCTION n1qlJS3 IF EXISTS")

	// Run the external JS UDF tests
	externalJSTest(qc, t)

	runStmt(qc, "DELETE FROM customer WHERE test_id = \"udf\"")
	runStmt(qc, "DROP PRIMARY INDEX ON customer")

	qc.ShutdownHttpServer()
}

func externalJSTest(qc *gsi.MockServer, t *testing.T) {

	// The JS UDFs will be loaded into a library named "lib1"
	url := "http://" + gsi.Auth_param + "@" + gsi.Query_CBS + "/evaluator/v1/libraries/lib1"
	client := &http.Client{}

	library :=
		`
		function external1(var1) {
			let var2 = 30;
			var selectquery = SELECT custId FROM customer WHERE test_id = "udf" AND age > $var1 AND age < $var2 ORDER BY custId;
			var rs = [];
			for (const row of selectquery) {
				rs.push(row);
				}
			selectquery.close()
			return rs;
		}

		// Function that executes another function
		function external2(var1) {
			var query = N1QL("EXECUTE FUNCTION externalJS1("+var1+")");
			var q = [];
			for (const row of query) {
				q.push(row);
			}
			query.close()
			return q;
		}

		// Function that performs a SELECT after a DML operation on the same bucket
		function external3() {
			// DML - UPDATE query
			var updateQuery = UPDATE customer SET externalChangeId = 1 WHERE test_id = "udf";
			updateQuery.close()
		
			// SELECT query
			var selectQuery = SELECT externalChangeId, COUNT(*) as count from customer WHERE test_id = "udf" GROUP BY externalChangeId;
			
			var q = [];
			for (const row of selectQuery) {
				q.push(row);
			}
		
			selectQuery.close();
		
			return q;
		}
	`

	// Load functions
	loadReq, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(library))
	loadResp, loadErr := client.Do(loadReq)

	if loadErr != nil || loadResp.StatusCode != http.StatusOK {
		t.Log("udf_test.go: Error creating and loading functions into library")
	}

	client.Do(loadReq)

	// run tests
	runMatch("case_external_js_udf_tests.json", false, true, qc, t)

	// Delete library
	delReq, _ := http.NewRequest(http.MethodDelete, url, nil)
	delResp, delErr := client.Do(delReq)

	if delErr != nil || delResp.StatusCode != http.StatusOK {
		t.Log("udf_test.go: Error deleting library")
	}

	client.Do(delReq)

	// Drop functions created in external JS UDF tests
	runStmt(qc, "DROP FUNCTION externalJS1 IF EXISTS")
	runStmt(qc, "DROP FUNCTION externalJS2 IF EXISTS")
	runStmt(qc, "DROP FUNCTION externalJS3 IF EXISTS")

}
