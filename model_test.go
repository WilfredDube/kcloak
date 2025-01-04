package kcloak_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/WilfredDube/kcloak"
)

func TestStringOrArray_Unmarshal(t *testing.T) {
	t.Parallel()
	jsonString := []byte("\"123\"")
	var dataString kcloak.StringOrArray
	err := json.Unmarshal(jsonString, &dataString)
	assert.NoErrorf(t, err, "Unmarshalling failed for json string: %s", jsonString)
	assert.Len(t, dataString, 1)
	assert.Equal(t, "123", dataString[0])

	jsonArray := []byte("[\"1\",\"2\",\"3\"]")
	var dataArray kcloak.StringOrArray
	err = json.Unmarshal(jsonArray, &dataArray)
	assert.NoError(t, err, "Unmarshalling failed for json array of strings: %s", jsonArray)
	assert.Len(t, dataArray, 3)
	assert.EqualValues(t, []string{"1", "2", "3"}, dataArray)
}

func TestStringOrArray_Marshal(t *testing.T) {
	t.Parallel()
	dataString := kcloak.StringOrArray{"123"}
	jsonString, err := json.Marshal(&dataString)
	assert.NoErrorf(t, err, "Marshaling failed for one string: %s", dataString)
	assert.Equal(t, "\"123\"", string(jsonString))

	dataArray := kcloak.StringOrArray{"1", "2", "3"}
	jsonArray, err := json.Marshal(&dataArray)
	assert.NoError(t, err, "Marshaling failed for array of strings: %s", dataArray)
	assert.Equal(t, "[\"1\",\"2\",\"3\"]", string(jsonArray))
}

func TestEnforcedString_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	type testData struct {
		In  []byte
		Out kcloak.EnforcedString
	}

	data := []testData{{
		In:  []byte(`"string value"`),
		Out: "string value",
	}, {
		In:  []byte(`"\"quoted string value\""`),
		Out: `"quoted string value"`,
	}, {
		In:  []byte(`true`),
		Out: "true",
	}, {
		In:  []byte(`42`),
		Out: "42",
	}, {
		In:  []byte(`{"foo": "bar"}`),
		Out: `{"foo": "bar"}`,
	}, {
		In:  []byte(`["foo"]`),
		Out: `["foo"]`,
	}}

	for _, d := range data {
		var val kcloak.EnforcedString
		err := json.Unmarshal(d.In, &val)
		assert.NoErrorf(t, err, "Unmarshalling failed with data: %v", d.In)
		assert.Equal(t, d.Out, val)
	}
}

func TestEnforcedString_MarshalJSON(t *testing.T) {
	t.Parallel()

	data := kcloak.EnforcedString("foo")
	jsonString, err := json.Marshal(&data)
	assert.NoErrorf(t, err, "Unmarshalling failed with data: %v", data)
	assert.Equal(t, `"foo"`, string(jsonString))
}

func TestGetQueryParams(t *testing.T) {
	t.Parallel()

	type TestParams struct {
		IntField    *int    `json:"int_field,string,omitempty"`
		StringField *string `json:"string_field,omitempty"`
		BoolField   *bool   `json:"bool_field,string,omitempty"`
	}

	params, err := kcloak.GetQueryParams(TestParams{})
	assert.NoError(t, err)
	assert.True(
		t,
		len(params) == 0,
		"Params must be empty, but got: %+v",
		params,
	)

	params, err = kcloak.GetQueryParams(TestParams{
		IntField:    kcloak.IntP(1),
		StringField: kcloak.StringP("fake"),
		BoolField:   kcloak.BoolP(true),
	})
	assert.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			"int_field":    "1",
			"string_field": "fake",
			"bool_field":   "true",
		},
		params,
	)

	params, err = kcloak.GetQueryParams(TestParams{
		StringField: kcloak.StringP("fake"),
		BoolField:   kcloak.BoolP(false),
	})
	assert.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			"string_field": "fake",
			"bool_field":   "false",
		},
		params,
	)
}

func TestParseAPIErrType(t *testing.T) {
	testCases := []struct {
		Name     string
		Error    error
		Expected kcloak.APIErrType
	}{
		{
			Name:     "nil error",
			Error:    nil,
			Expected: kcloak.APIErrTypeUnknown,
		},
		{
			Name:     "invalid grant",
			Error:    errors.New("something something invalid_grant something"),
			Expected: kcloak.APIErrTypeInvalidGrant,
		},
		{
			Name:     "other error",
			Error:    errors.New("something something unsupported_grant_type something"),
			Expected: kcloak.APIErrTypeUnknown,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := kcloak.ParseAPIErrType(testCase.Error)
			if result != testCase.Expected {
				t.Fatalf("expected %s but received %s", testCase.Expected, result)
			}
		})
	}
}

func TestStringer(t *testing.T) {
	// nested structs
	actions := []string{"someAction", "anotherAction"}
	access := kcloak.AccessRepresentation{
		Manage:      kcloak.BoolP(true),
		Impersonate: kcloak.BoolP(false),
	}
	v := kcloak.PermissionTicketDescriptionRepresentation{
		ID:               kcloak.StringP("someID"),
		CreatedTimeStamp: kcloak.Int64P(1607702613),
		Enabled:          kcloak.BoolP(true),
		RequiredActions:  &actions,
		Access:           &access,
	}

	str := v.String()

	expectedStr := `{
	"id": "someID",
	"createdTimestamp": 1607702613,
	"enabled": true,
	"requiredActions": [
		"someAction",
		"anotherAction"
	],
	"access": {
		"impersonate": false,
		"manage": true
	}
}`

	assert.Equal(t, expectedStr, str)

	// nested arrays
	config := make(map[string]string)
	config["bar"] = "foo"
	config["ping"] = "pong"

	pmappers := []kcloak.ProtocolMapperRepresentation{
		{
			Name:   kcloak.StringP("someMapper"),
			Config: &config,
		},
	}
	clients := []kcloak.Client{
		{
			Name:            kcloak.StringP("someClient"),
			ProtocolMappers: &pmappers,
		},
		{
			Name: kcloak.StringP("AnotherClient"),
		},
	}

	realmRep := kcloak.RealmRepresentation{
		DisplayName: kcloak.StringP("someRealm"),
		Clients:     &clients,
	}

	str = realmRep.String()
	expectedStr = `{
	"clients": [
		{
			"name": "someClient",
			"protocolMappers": [
				{
					"config": {
						"bar": "foo",
						"ping": "pong"
					},
					"name": "someMapper"
				}
			]
		},
		{
			"name": "AnotherClient"
		}
	],
	"displayName": "someRealm"
}`
	assert.Equal(t, expectedStr, str)
}

type Stringable interface {
	String() string
}

func TestStringerOmitEmpty(t *testing.T) {
	customs := []Stringable{
		&kcloak.CertResponseKey{},
		&kcloak.CertResponse{},
		&kcloak.IssuerResponse{},
		&kcloak.ResourcePermission{},
		&kcloak.PermissionResource{},
		&kcloak.PermissionScope{},
		&kcloak.IntroSpectTokenResult{},
		&kcloak.User{},
		&kcloak.SetPasswordRequest{},
		&kcloak.Component{},
		&kcloak.KeyStoreConfig{},
		&kcloak.ActiveKeys{},
		&kcloak.Key{},
		&kcloak.Attributes{},
		&kcloak.Access{},
		&kcloak.UserGroup{},
		&kcloak.ExecuteActionsEmail{},
		&kcloak.Group{},
		&kcloak.GroupsCount{},
		&kcloak.GetGroupsParams{},
		&kcloak.CompositesRepresentation{},
		&kcloak.Role{},
		&kcloak.GetRoleParams{},
		&kcloak.ClientMappingsRepresentation{},
		&kcloak.MappingsRepresentation{},
		&kcloak.ClientScope{},
		&kcloak.ClientScopeAttributes{},
		&kcloak.ProtocolMappers{},
		&kcloak.ProtocolMappersConfig{},
		&kcloak.Client{},
		&kcloak.ResourceServerRepresentation{},
		&kcloak.RoleDefinition{},
		&kcloak.PolicyRepresentation{},
		&kcloak.RolePolicyRepresentation{},
		&kcloak.JSPolicyRepresentation{},
		&kcloak.ClientPolicyRepresentation{},
		&kcloak.TimePolicyRepresentation{},
		&kcloak.UserPolicyRepresentation{},
		&kcloak.AggregatedPolicyRepresentation{},
		&kcloak.GroupPolicyRepresentation{},
		&kcloak.GroupDefinition{},
		&kcloak.ResourceRepresentation{},
		&kcloak.ResourceOwnerRepresentation{},
		&kcloak.ScopeRepresentation{},
		&kcloak.ProtocolMapperRepresentation{},
		&kcloak.UserInfoAddress{},
		&kcloak.UserInfo{},
		&kcloak.RolesRepresentation{},
		&kcloak.RealmRepresentation{},
		&kcloak.MultiValuedHashMap{},
		&kcloak.TokenOptions{},
		&kcloak.UserSessionRepresentation{},
		&kcloak.SystemInfoRepresentation{},
		&kcloak.MemoryInfoRepresentation{},
		&kcloak.ServerInfoRepresentation{},
		&kcloak.FederatedIdentityRepresentation{},
		&kcloak.IdentityProviderRepresentation{},
		&kcloak.GetResourceParams{},
		&kcloak.GetScopeParams{},
		&kcloak.GetPolicyParams{},
		&kcloak.GetPermissionParams{},
		&kcloak.GetUsersByRoleParams{},
		&kcloak.PermissionRepresentation{},
		&kcloak.CreatePermissionTicketParams{},
		&kcloak.PermissionTicketDescriptionRepresentation{},
		&kcloak.AccessRepresentation{},
		&kcloak.PermissionTicketResponseRepresentation{},
		&kcloak.PermissionTicketRepresentation{},
		&kcloak.PermissionTicketPermissionRepresentation{},
		&kcloak.PermissionGrantParams{},
		&kcloak.PermissionGrantResponseRepresentation{},
		&kcloak.GetUserPermissionParams{},
		&kcloak.ResourcePolicyRepresentation{},
		&kcloak.GetResourcePoliciesParams{},
		&kcloak.CredentialRepresentation{},
		&kcloak.GetUsersParams{},
		&kcloak.GetComponentsParams{},
		&kcloak.GetClientsParams{},
		&kcloak.RequestingPartyTokenOptions{},
		&kcloak.RequestingPartyPermission{},
		&kcloak.GetClientUserSessionsParams{},
		&kcloak.GetOrganizationsParams{},
		&kcloak.OrganizationDomainRepresentation{},
		&kcloak.OrganizationRepresentation{},
	}

	for _, custom := range customs {
		assert.Equal(t, "{}", custom.String())
	}
}
