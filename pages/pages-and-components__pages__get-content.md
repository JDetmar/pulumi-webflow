# Get Page Content

GET https://api.webflow.com/v2/pages/{page_id}/dom

Get text and component instance content from a static page.LocalizationRequired scope | `pages:read`

Reference: https://developers.webflow.com/data/reference/pages-and-components/pages/get-content

## Authentication

- `Authorization` header (bearer token, required) — Bearer authentication of the form `Bearer <token>`, where token is your auth token.

## Request

### Path parameters

- `page_id` (string, required) — Unique identifier for a Page

### Query parameters

- `localeId` (string, optional) — Unique identifier for a specific Locale. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)
- `limit` (integer, optional) — Maximum number of records to be returned (max limit: 100)
- `offset` (integer, optional) — Offset used for pagination if the results have more than limit records
- `translatable` (string, optional) — Unique identifier for the secondary Locale you're translating **into**. Returns only content that hasn't been excluded from translation for that locale. This is independent of `localeId`, which selects which version of the content is returned. To fetch the source text to translate, request the primary locale's content and set `translatable` to the locale you're translating into: `?localeId={primary locale id}&translatable={target locale id}` Only exclusion rules scoped to manual translation are respected — rules scoped only to automatic translation don't affect this parameter's response. Omitting `translatable` returns the same response as if this parameter didn't exist. The value must be the id of one of the site's secondary locales — the primary locale id, or any other value, returns a `400` error. Requires translation exclusions to be enabled for the site; if they aren't, the request returns a `403` error. [Learn more about localization.](/data/v2.0.0/docs/working-with-localization)

## Response

### 200

Request was successful

- `pageId` (string, optional) — Page ID
- `branchId` (string, optional, nullable) — The unique identifier of a [specific page branch.](https://help.webflow.com/hc/en-us/articles/33961355506195-Page-branching)
- `nodes` (list of PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItems, optional)
- `pagination` (PagesPageIdDomGetResponsesContentApplicationJsonSchemaPagination, optional) — Pagination object
- `lastUpdated` (string, optional, nullable) — The date the page dom was most recently updated

## Types

### PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItems

A generic representation of a content element within the Document Object Model (DOM). Each node has a unique identifier and a specific type that determines its content structure and attributes.

- Text Node DOM: Text Node DOM
- Image Node: Image Node
- Component Node: Component Node
- Text Input Node: Text Input Node
- Select Node: Select Node
- Submit Button Node: Submit Button Node
- Search Button Node: Search Button Node

### PagesPageIdDomGetResponsesContentApplicationJsonSchemaPagination

Pagination object

- `limit` (integer, required) — The limit used for pagination
- `offset` (integer, required) — The offset used for pagination
- `total` (integer, required) — The total number of records

### PagesPageIdDomGetResponsesContentApplicationJsonSchemaDetailsItems

- string
- map from string to any

### Text Node DOM

Represents text content within the DOM. It contains both the raw text and its HTML representation. Additional attributes can be associated with the text for styling or other purposes.

- `id` (string, required) — Node UUID
- `type` (enum, required, default: text) — The type of the node
  - Allowed values: `text`
- `text` (PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf0Text, required) — The text content of the node
- `attributes` (map from string to string, optional) — The custom attributes of the node

### Image Node

Represents an image within the DOM. It contains details about the image, such as its alternative text (alt) for accessibility and an asset identifier for fetching the actual image resource. Additional attributes can be associated with the image for styling or other purposes.

- `id` (string, required) — Node UUID
- `type` (enum, required, default: image) — The type of the node
  - Allowed values: `image`
- `image` (PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf1Image, required) — The image details of the node
- `attributes` (map from string to string, optional) — The custom attributes of the node

### Component Node

Represents a component instance within the DOM. It contains details about the component instance, such as its type and properties.

- `id` (string, required) — The unique identifier of the component instance node
- `type` (enum, required, default: component-instance) — The type of the node
  - Allowed values: `component-instance`
- `componentId` (string, required) — The unique identifier of the component
- `propertyOverrides` (list of PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf2PropertyOverridesItems, required) — List of component properties with overrides for a component instance.

### Text Input Node

Represents text input and textarea elements within the DOM. It contains the placeholder text in the input. Additional attributes can be associated with the text for styling or other purposes.

- `id` (string, required) — Node UUID
- `type` (enum, required, default: text-input) — The type of the node
  - Allowed values: `text-input`
- `placeholder` (string, required) — The placeholder text of the input node
- `attributes` (map from string to string, optional) — The custom attributes of the node

### Select Node

Represents select elements within the DOM. It contains the list of choices in the select. Additional attributes can be associated with the text for styling or other purposes.

- `id` (string, required) — Node UUID
- `type` (enum, required, default: select) — The type of the node
  - Allowed values: `select`
- `choices` (list of PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf4ChoicesItems, required) — The list of choices in this select node.
- `attributes` (map from string to string, optional) — The custom attributes of the node

### Submit Button Node

Represents submit button elements within the DOM. It contains the text and waiting text of the button. Additional attributes can be associated with the text for styling or other purposes.

- `id` (string, required) — Node UUID
- `type` (enum, required, default: submit-button) — The type of the node
  - Allowed values: `submit-button`
- `value` (string, required) — The text content of the submit button.
- `waitingText` (string, required) — The text to show while the form is submitting.
- `attributes` (map from string to string, optional) — The custom attributes of the node

### Search Button Node

Represents search button elements within the DOM. It contains the text of the button. Additional attributes can be associated with the text for styling or other purposes.

- `id` (string, required) — Node UUID
- `type` (enum, required, default: search-button) — The type of the node
  - Allowed values: `search-button`
- `value` (string, required) — The text content of the search button.
- `attributes` (map from string to string, optional) — The custom attributes of the node

### PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf0Text

The text content of the node

- `html` (string, optional, nullable) — The HTML content of the text node.
- `text` (string, optional, nullable) — The raw text content of the text node.

### PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf1Image

The image details of the node

- `alt` (string, optional, nullable)
- `assetId` (string, optional)

### PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf2PropertyOverridesItems

- Component Property Text: Component Property Text

### PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf4ChoicesItems

- `value` (string, required) — The value of the choice when selected.
- `text` (string, required) — The text to display for the choice.

### Component Property Text

A text-based component property containing raw text and HTML representation.

- `propertyId` (string, required) — The ID of the property.
- `type` (enum, required) — The type of the property.
  - Allowed values: `Plain Text`, `Rich Text`, `Alt Text`
- `label` (string, required) — The label of the property in the UI.
- `text` (PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf2PropertyOverridesItemsOneOf0Text, required) — Represents text content within the DOM. It contains both the raw text and its HTML representation.

### PagesPageIdDomGetResponsesContentApplicationJsonSchemaNodesItemsOneOf2PropertyOverridesItemsOneOf0Text

Represents text content within the DOM. It contains both the raw text and its HTML representation.

- `html` (string, optional, nullable) — The HTML content of the text node.
- `text` (string, optional, nullable) — The raw text content of the text node.
