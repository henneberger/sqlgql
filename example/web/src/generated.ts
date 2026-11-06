import { JSONValue } from './json';
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  BigInt: { input: string; output: string; }
  JSON: { input: JSONValue; output: JSONValue; }
  UUID: { input: string; output: string; }
};

export type Customer = {
  __typename?: 'Customer';
  customerItems: Array<CustomerItems>;
  email?: Maybe<Scalars['String']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  name?: Maybe<Scalars['String']['output']>;
};


export type CustomerCustomerItemsArgs = {
  limit: Scalars['Int']['input'];
};

export type CustomerItems = {
  __typename?: 'CustomerItems';
  customerId?: Maybe<Scalars['Int']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  label?: Maybe<Scalars['String']['output']>;
  quantity?: Maybe<Scalars['Int']['output']>;
};

export type CustomerPage = {
  __typename?: 'CustomerPage';
  email?: Maybe<Scalars['String']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  name?: Maybe<Scalars['String']['output']>;
  pageItems: PageItemsConnection;
};


export type CustomerPagePageItemsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};

export type CustomerPageConnection = {
  __typename?: 'CustomerPageConnection';
  edges: Array<CustomerPageEdge>;
  pageInfo: PageInfo;
};

export type CustomerPageEdge = {
  __typename?: 'CustomerPageEdge';
  cursor: Scalars['String']['output'];
  node: CustomerPage;
};

export type Customers = {
  __typename?: 'Customers';
  email?: Maybe<Scalars['String']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  items: Array<Items>;
  name?: Maybe<Scalars['String']['output']>;
};


export type CustomersItemsArgs = {
  limit: Scalars['Int']['input'];
};

export type Items = {
  __typename?: 'Items';
  customerId?: Maybe<Scalars['Int']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  label?: Maybe<Scalars['String']['output']>;
  quantity?: Maybe<Scalars['Int']['output']>;
};

export type NumericDemo = {
  __typename?: 'NumericDemo';
  big?: Maybe<Scalars['BigInt']['output']>;
  bigs?: Maybe<Array<Maybe<Scalars['BigInt']['output']>>>;
  double?: Maybe<Scalars['Float']['output']>;
  doubles?: Maybe<Array<Maybe<Scalars['Float']['output']>>>;
  optionalBig?: Maybe<Scalars['BigInt']['output']>;
  optionalDouble?: Maybe<Scalars['Float']['output']>;
  optionalReal?: Maybe<Scalars['Float']['output']>;
  real?: Maybe<Scalars['Float']['output']>;
  reals?: Maybe<Array<Maybe<Scalars['Float']['output']>>>;
  special?: Maybe<Scalars['Float']['output']>;
};

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
  hasPreviousPage: Scalars['Boolean']['output'];
  startCursor?: Maybe<Scalars['String']['output']>;
};

export type PageItems = {
  __typename?: 'PageItems';
  customerId?: Maybe<Scalars['Int']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  label?: Maybe<Scalars['String']['output']>;
  quantity?: Maybe<Scalars['Int']['output']>;
};

export type PageItemsConnection = {
  __typename?: 'PageItemsConnection';
  edges: Array<PageItemsEdge>;
  pageInfo: PageInfo;
};

export type PageItemsEdge = {
  __typename?: 'PageItemsEdge';
  cursor: Scalars['String']['output'];
  node: PageItems;
};

export type Query = {
  __typename?: 'Query';
  customer?: Maybe<Customer>;
  customerPage: CustomerPageConnection;
  customers: Array<Customers>;
  numericDemo?: Maybe<NumericDemo>;
  typeDemo?: Maybe<TypeDemo>;
};


export type QueryCustomerArgs = {
  id: Scalars['Int']['input'];
};


export type QueryCustomerPageArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryCustomersArgs = {
  limit: Scalars['Int']['input'];
};


export type QueryNumericDemoArgs = {
  big: Scalars['BigInt']['input'];
  bigs: Array<InputMaybe<Scalars['BigInt']['input']>>;
  double: Scalars['Float']['input'];
  doubles: Array<InputMaybe<Scalars['Float']['input']>>;
  real: Scalars['Float']['input'];
  reals: Array<InputMaybe<Scalars['Float']['input']>>;
  special: Scalars['String']['input'];
};


export type QueryTypeDemoArgs = {
  counts: Array<InputMaybe<Scalars['Int']['input']>>;
  document: Scalars['JSON']['input'];
  documents: Array<InputMaybe<Scalars['JSON']['input']>>;
  flags: Array<InputMaybe<Scalars['Boolean']['input']>>;
  id: Scalars['UUID']['input'];
  ids: Array<InputMaybe<Scalars['UUID']['input']>>;
  labels: Array<InputMaybe<Scalars['String']['input']>>;
  raw: Scalars['JSON']['input'];
  scores: Array<InputMaybe<Scalars['Float']['input']>>;
};

export type TypeDemo = {
  __typename?: 'TypeDemo';
  counts?: Maybe<Array<Maybe<Scalars['Int']['output']>>>;
  document?: Maybe<Scalars['JSON']['output']>;
  documents?: Maybe<Array<Maybe<Scalars['JSON']['output']>>>;
  emptyCounts?: Maybe<Array<Maybe<Scalars['Int']['output']>>>;
  flags?: Maybe<Array<Maybe<Scalars['Boolean']['output']>>>;
  id?: Maybe<Scalars['UUID']['output']>;
  ids?: Maybe<Array<Maybe<Scalars['UUID']['output']>>>;
  labels?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  optionalDocument?: Maybe<Scalars['JSON']['output']>;
  optionalId?: Maybe<Scalars['UUID']['output']>;
  optionalLabels?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  raw?: Maybe<Scalars['JSON']['output']>;
  scores?: Maybe<Array<Maybe<Scalars['Float']['output']>>>;
};

export type CustomerQueryVariables = Exact<{
  Query_customer_id: Scalars['Int']['input'];
  Query_customer_customerItems_limit: Scalars['Int']['input'];
}>;


export type CustomerQuery = { __typename?: 'Query', customer?: { __typename: 'Customer', id?: number | null, name?: string | null, email?: string | null, customerItems: Array<{ __typename: 'CustomerItems', customerId?: number | null, id?: number | null, label?: string | null, quantity?: number | null }> } | null };

export type CustomerPageQueryVariables = Exact<{
  Query_customerPage_first?: InputMaybe<Scalars['Int']['input']>;
  Query_customerPage_after?: InputMaybe<Scalars['String']['input']>;
  Query_customerPage_last?: InputMaybe<Scalars['Int']['input']>;
  Query_customerPage_before?: InputMaybe<Scalars['String']['input']>;
  Query_customerPage_pageItems_first?: InputMaybe<Scalars['Int']['input']>;
  Query_customerPage_pageItems_after?: InputMaybe<Scalars['String']['input']>;
  Query_customerPage_pageItems_last?: InputMaybe<Scalars['Int']['input']>;
  Query_customerPage_pageItems_before?: InputMaybe<Scalars['String']['input']>;
}>;


export type CustomerPageQuery = { __typename?: 'Query', customerPage: { __typename: 'CustomerPageConnection', pageInfo: { __typename: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null }, edges: Array<{ __typename: 'CustomerPageEdge', cursor: string, node: { __typename: 'CustomerPage', id?: number | null, name?: string | null, email?: string | null, pageItems: { __typename: 'PageItemsConnection', pageInfo: { __typename: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null }, edges: Array<{ __typename: 'PageItemsEdge', cursor: string, node: { __typename: 'PageItems', customerId?: number | null, id?: number | null, label?: string | null, quantity?: number | null } }> } } }> } };

export type CustomersQueryVariables = Exact<{
  Query_customers_limit: Scalars['Int']['input'];
  Query_customers_items_limit: Scalars['Int']['input'];
}>;


export type CustomersQuery = { __typename?: 'Query', customers: Array<{ __typename: 'Customers', id?: number | null, name?: string | null, email?: string | null, items: Array<{ __typename: 'Items', customerId?: number | null, id?: number | null, label?: string | null, quantity?: number | null }> }> };

export type NumericDemoQueryVariables = Exact<{
  Query_numericDemo_big: Scalars['BigInt']['input'];
  Query_numericDemo_real: Scalars['Float']['input'];
  Query_numericDemo_double: Scalars['Float']['input'];
  Query_numericDemo_bigs: Array<InputMaybe<Scalars['BigInt']['input']>> | InputMaybe<Scalars['BigInt']['input']>;
  Query_numericDemo_reals: Array<InputMaybe<Scalars['Float']['input']>> | InputMaybe<Scalars['Float']['input']>;
  Query_numericDemo_doubles: Array<InputMaybe<Scalars['Float']['input']>> | InputMaybe<Scalars['Float']['input']>;
  Query_numericDemo_special: Scalars['String']['input'];
}>;


export type NumericDemoQuery = { __typename?: 'Query', numericDemo?: { __typename: 'NumericDemo', big?: string | null, real?: number | null, double?: number | null, bigs?: Array<string | null> | null, reals?: Array<number | null> | null, doubles?: Array<number | null> | null, optionalBig?: string | null, optionalReal?: number | null, optionalDouble?: number | null, special?: number | null } | null };

export type TypeDemoQueryVariables = Exact<{
  Query_typeDemo_id: Scalars['UUID']['input'];
  Query_typeDemo_document: Scalars['JSON']['input'];
  Query_typeDemo_raw: Scalars['JSON']['input'];
  Query_typeDemo_ids: Array<InputMaybe<Scalars['UUID']['input']>> | InputMaybe<Scalars['UUID']['input']>;
  Query_typeDemo_labels: Array<InputMaybe<Scalars['String']['input']>> | InputMaybe<Scalars['String']['input']>;
  Query_typeDemo_counts: Array<InputMaybe<Scalars['Int']['input']>> | InputMaybe<Scalars['Int']['input']>;
  Query_typeDemo_flags: Array<InputMaybe<Scalars['Boolean']['input']>> | InputMaybe<Scalars['Boolean']['input']>;
  Query_typeDemo_scores: Array<InputMaybe<Scalars['Float']['input']>> | InputMaybe<Scalars['Float']['input']>;
  Query_typeDemo_documents: Array<InputMaybe<Scalars['JSON']['input']>> | InputMaybe<Scalars['JSON']['input']>;
}>;


export type TypeDemoQuery = { __typename?: 'Query', typeDemo?: { __typename: 'TypeDemo', id?: string | null, document?: JSONValue | null, raw?: JSONValue | null, ids?: Array<string | null> | null, labels?: Array<string | null> | null, counts?: Array<number | null> | null, flags?: Array<boolean | null> | null, scores?: Array<number | null> | null, documents?: Array<JSONValue | null> | null, optionalId?: string | null, optionalDocument?: JSONValue | null, optionalLabels?: Array<string | null> | null, emptyCounts?: Array<number | null> | null } | null };


export const CustomerDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Customer"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customer_id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customer_customerItems_limit"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"customer"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customer_id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"customerItems"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customer_customerItems_limit"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"customerId"}},{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"label"}},{"kind":"Field","name":{"kind":"Name","value":"quantity"}}]}}]}}]}}]} as unknown as DocumentNode<CustomerQuery, CustomerQueryVariables>;
export const CustomerPageDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"CustomerPage"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_first"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_after"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_last"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_before"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_pageItems_first"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_pageItems_after"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_pageItems_last"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_pageItems_before"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"customerPage"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"first"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_first"}}},{"kind":"Argument","name":{"kind":"Name","value":"after"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_after"}}},{"kind":"Argument","name":{"kind":"Name","value":"last"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_last"}}},{"kind":"Argument","name":{"kind":"Name","value":"before"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_before"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"pageInfo"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"hasNextPage"}},{"kind":"Field","name":{"kind":"Name","value":"hasPreviousPage"}},{"kind":"Field","name":{"kind":"Name","value":"startCursor"}},{"kind":"Field","name":{"kind":"Name","value":"endCursor"}}]}},{"kind":"Field","name":{"kind":"Name","value":"edges"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"cursor"}},{"kind":"Field","name":{"kind":"Name","value":"node"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"pageItems"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"first"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_pageItems_first"}}},{"kind":"Argument","name":{"kind":"Name","value":"after"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_pageItems_after"}}},{"kind":"Argument","name":{"kind":"Name","value":"last"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_pageItems_last"}}},{"kind":"Argument","name":{"kind":"Name","value":"before"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customerPage_pageItems_before"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"pageInfo"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"hasNextPage"}},{"kind":"Field","name":{"kind":"Name","value":"hasPreviousPage"}},{"kind":"Field","name":{"kind":"Name","value":"startCursor"}},{"kind":"Field","name":{"kind":"Name","value":"endCursor"}}]}},{"kind":"Field","name":{"kind":"Name","value":"edges"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"cursor"}},{"kind":"Field","name":{"kind":"Name","value":"node"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"customerId"}},{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"label"}},{"kind":"Field","name":{"kind":"Name","value":"quantity"}}]}}]}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<CustomerPageQuery, CustomerPageQueryVariables>;
export const CustomersDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Customers"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customers_limit"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_customers_items_limit"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"customers"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customers_limit"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"name"}},{"kind":"Field","name":{"kind":"Name","value":"email"}},{"kind":"Field","name":{"kind":"Name","value":"items"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"limit"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_customers_items_limit"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"customerId"}},{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"label"}},{"kind":"Field","name":{"kind":"Name","value":"quantity"}}]}}]}}]}}]} as unknown as DocumentNode<CustomersQuery, CustomersQueryVariables>;
export const NumericDemoDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"NumericDemo"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_big"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BigInt"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_real"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Float"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_double"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Float"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_bigs"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"BigInt"}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_reals"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Float"}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_doubles"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Float"}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_special"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"numericDemo"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"big"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_big"}}},{"kind":"Argument","name":{"kind":"Name","value":"real"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_real"}}},{"kind":"Argument","name":{"kind":"Name","value":"double"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_double"}}},{"kind":"Argument","name":{"kind":"Name","value":"bigs"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_bigs"}}},{"kind":"Argument","name":{"kind":"Name","value":"reals"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_reals"}}},{"kind":"Argument","name":{"kind":"Name","value":"doubles"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_doubles"}}},{"kind":"Argument","name":{"kind":"Name","value":"special"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_numericDemo_special"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"big"}},{"kind":"Field","name":{"kind":"Name","value":"real"}},{"kind":"Field","name":{"kind":"Name","value":"double"}},{"kind":"Field","name":{"kind":"Name","value":"bigs"}},{"kind":"Field","name":{"kind":"Name","value":"reals"}},{"kind":"Field","name":{"kind":"Name","value":"doubles"}},{"kind":"Field","name":{"kind":"Name","value":"optionalBig"}},{"kind":"Field","name":{"kind":"Name","value":"optionalReal"}},{"kind":"Field","name":{"kind":"Name","value":"optionalDouble"}},{"kind":"Field","name":{"kind":"Name","value":"special"}}]}}]}}]} as unknown as DocumentNode<NumericDemoQuery, NumericDemoQueryVariables>;
export const TypeDemoDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"TypeDemo"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UUID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_document"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"JSON"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_raw"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"JSON"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_ids"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"UUID"}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_labels"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_counts"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Int"}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_flags"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Boolean"}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_scores"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"Float"}}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_documents"}},"type":{"kind":"NonNullType","type":{"kind":"ListType","type":{"kind":"NamedType","name":{"kind":"Name","value":"JSON"}}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"typeDemo"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_id"}}},{"kind":"Argument","name":{"kind":"Name","value":"document"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_document"}}},{"kind":"Argument","name":{"kind":"Name","value":"raw"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_raw"}}},{"kind":"Argument","name":{"kind":"Name","value":"ids"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_ids"}}},{"kind":"Argument","name":{"kind":"Name","value":"labels"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_labels"}}},{"kind":"Argument","name":{"kind":"Name","value":"counts"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_counts"}}},{"kind":"Argument","name":{"kind":"Name","value":"flags"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_flags"}}},{"kind":"Argument","name":{"kind":"Name","value":"scores"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_scores"}}},{"kind":"Argument","name":{"kind":"Name","value":"documents"},"value":{"kind":"Variable","name":{"kind":"Name","value":"Query_typeDemo_documents"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"__typename"}},{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"document"}},{"kind":"Field","name":{"kind":"Name","value":"raw"}},{"kind":"Field","name":{"kind":"Name","value":"ids"}},{"kind":"Field","name":{"kind":"Name","value":"labels"}},{"kind":"Field","name":{"kind":"Name","value":"counts"}},{"kind":"Field","name":{"kind":"Name","value":"flags"}},{"kind":"Field","name":{"kind":"Name","value":"scores"}},{"kind":"Field","name":{"kind":"Name","value":"documents"}},{"kind":"Field","name":{"kind":"Name","value":"optionalId"}},{"kind":"Field","name":{"kind":"Name","value":"optionalDocument"}},{"kind":"Field","name":{"kind":"Name","value":"optionalLabels"}},{"kind":"Field","name":{"kind":"Name","value":"emptyCounts"}}]}}]}}]} as unknown as DocumentNode<TypeDemoQuery, TypeDemoQueryVariables>;