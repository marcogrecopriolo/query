#!/bin/bash

#  Copyright 2023-Present Couchbase, Inc.
#
#  Use of this software is governed by the Business Source License included
#  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
#  in that file, in accordance with the Business Source License, use of this
#  software will be governed by the Apache License, Version 2.0, included in
#  the file licenses/APL2.txt.

# This scripts generates a go-lang map object containing the grammar used by the Query engine

if [ $# != 1 ]
then
  echo "Missing argument to $0"
  exit 1
fi

BASEPATH=$1
FILE=${BASEPATH}/shell/cbq/command/syntax_data.go

if [ ! -f ${BASEPATH}/parser/n1ql/n1ql.y ]
then
  echo "Invalid base path: ${BASEPATH}"
  exit 1
fi

AC='
BEGIN \
{
  terms[0]=""
  delete terms[0]
  error=0
  lastStart=1
}
NF!=1 { next }
length(terms)==0 \
{
  terms[0]=fixTerm($NF)
  getline
  while (NF==0) if (getline!=1) { error=1; exit }
  if ($NF!=":") { print "ERROR at token "NR" - \""$0"\" unexpected"; error=1; exit }
  next
}
/\// \
{
  getline
  while ($NF!="/") if (getline!=1) { error=1; exit }
  next
}
/\{/ \
{
  getline
  while ($NF!="}") if (getline!=1) { error=1; exit }
  next
}
/%/ \
{
  getline
  getline
  next
}
/;/ \
{
  checkAndSkip()
  if (terms[length(terms)-1]=="|") delete terms[length(terms)-1]
  if (length(terms)>1&&!excludeRule(terms[0])) printRule(terms)
  delete terms
  lastStart=1
  next
}
/\|/ \
{
  n=length(terms)
  if (n==1||terms[n-1]=="|") next
  if (checkAndSkip()==1) next
  lastStart=n+1
}
{
  terms[length(terms)]=fixTerm($NF)
}
END \
{
  if (error!=0) exit;
  print "}"
}
function printRule(terms) \
{
  if (length(terms)<2||(length(terms)==2&&terms[0]=="["terms[1]"]")) return
  comma=0
  for (i=0;i<length(terms);i++)
  {
    if (i==0)
    {
      printf("\t\"%s\": [][]string{\n\t\t[]string{",terms[0])
    }
    else if (terms[i]=="|")
    {
      printf("},\n\t\t[]string{")
      comma=0
    }
    else
    {
      if (comma==1) printf(",")
      printf("\"%s\"",terms[i])
      comma=1
    }
  }
  printf("},\n\t},\n")
}
function fixTerm(t) \
{
  if (t=="stmt_body")
  {
    return "statements"
  }
  else if (t=="ident_or_default")
  {
    return "<identifier>"
  }  
  else if (t=="IDENT")
  {
    return "<identifier>"
  }  
  else if (t=="IDENT_ICASE")
  {
    return "<identifier>i"
  }  
  else if (t=="NAMESPACE_ID")
  {
    return "<namespace-identifier>"
  }  
  else if (t=="expr")
  {
    return "expression"
  }  
  else if (t=="OPTIM_HINTS")
  {
    return "/*OPTIM_HINTS*/"
  }  
  else if (t=="hints_input")
  {
    return "OPTIM_HINTS"
  }
  gsub("stmt","statement",t)
  if (index(t,"opt_")==1) t = "["substr(t,5)"]"
  return t
}
function excludeRule(r) \
{
  if (r=="input"||r=="expr_input"||r=="<identifier>"||r=="[trailer]")
  {
    return 1
  }
  if (index(r,"$")!=0||substr(r,1,1)=="_") return 1
  return 0
}
function checkAndSkip() \
{
  drop=0
  n=length(terms)
  for (i=lastStart;i<n;i++) if (substr(terms[i],1,1)=="_") { drop=1; break }
  if (drop==1)
  {
    for (i=lastStart;i<n;i++) delete terms[i]
    return 1
  }
  return 0
}
'

cat - << EOF > "${FILE}"
//  Copyright 2023-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

// WARNING: This file is generated by the build process. DO NOT EDIT.

package command

var statement_syntax = map[string][][]string{
EOF

sed -n '/^%%/,$ p' ${BASEPATH}/parser/n1ql/n1ql.y \
  | sed -e '/^%%/d;/^[*/}{[:space:]]/d;/^$/d;s/[[:space:]]*\([A-Za-z_][A-Za-z0-9_]*\)[[:space:]]*/\1\n/g' \
  | sed -e '/^[^A-Za-z_]/s/\([:;|/*]\)/\n\1\n/g' \
  | awk "${AC}" >> "${FILE}"

go fmt ${FILE} 2>/dev/null
if [ $? -ne 0 ]
then
  rm -f ${FILE}
  echo "ERROR: Failed to generate syntax help data"
  exit 1
fi
