//  Copyright (c) 2019 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

// +build enterprise

package plan

import (
	"encoding/json"

	"github.com/couchbase/query-ee/indexadvisor/iaplan"
	"github.com/couchbase/query/expression"
)

type IndexAdvice struct {
	readonly
	adviceInfos iaplan.IndexAdviceInfos
}

func NewIndexAdvice(queryInfos map[expression.HasExpressions]*iaplan.QueryInfo) *IndexAdvice {
	rv := &IndexAdvice{}
	rv.adviceInfos = make(iaplan.IndexAdviceInfos, 0, len(queryInfos))
	qLen := len(queryInfos)
	cnt := 0
	for _, v := range queryInfos {
		cnt += 1
		adviceInfo := iaplan.NewIndexAdviceInfo(v.GetCurIndexes(), v.GetUncoverIndexes(), v.GetCoverIndexes(), v.IsKeyspaceFound())
		//MB-35353: get rid of multiple empty entryies when there are subquries
		if qLen == 1 || (qLen > 1 && (!adviceInfo.IndexesEmpty() || len(rv.adviceInfos) == 0 && cnt == qLen)) {
			rv.adviceInfos = append(rv.adviceInfos, adviceInfo)
		}
	}
	return rv
}

func (this *IndexAdvice) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitIndexAdvice(this)
}

func (this *IndexAdvice) New() Operator {
	return &IndexAdvice{}
}

func (this *IndexAdvice) Operator() Operator {
	return this
}

func (this *IndexAdvice) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.MarshalBase(nil))
}

func (this *IndexAdvice) MarshalBase(f func(map[string]interface{})) map[string]interface{} {
	r := map[string]interface{}{"#operator": "IndexAdvice"}
	r["adviseinfo"] = this.adviceInfos

	if f != nil {
		f(r)
	}
	return r
}

func (this *IndexAdvice) UnmarshalJSON(body []byte) error {
	var _unmarshalled struct {
		_           string            `json:"#operator"`
		AdviceInfos []json.RawMessage `json:"adviseinfo"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return err
	}

	if len(_unmarshalled.AdviceInfos) > 0 {
		this.adviceInfos = make(iaplan.IndexAdviceInfos, len(_unmarshalled.AdviceInfos))
		for _, v := range _unmarshalled.AdviceInfos {
			r := &iaplan.IndexAdviceInfo{}
			err = r.UnmarshalJSON(v)
			if err != nil {
				return err
			}
			this.adviceInfos = append(this.adviceInfos, r)
		}
	}

	return nil
}
