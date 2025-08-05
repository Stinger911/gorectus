import collectionsService, { Field, CreateFieldRequest, UpdateFieldRequest, FieldSchema } from './collectionsService';

// Field interfaces and types based on Directus patterns
export interface FieldInterface {
  id: string;
  name: string;
  icon: string;
  description: string;
  component: string;
  types: string[];
  groups: string[];
  options?: any;
  system?: boolean;
}

export interface FieldType {
  id: string;
  name: string;
  icon: string;
  description: string;
  sql: string;
  length?: boolean;
  decimal?: boolean;
  nullable?: boolean;
  default?: boolean;
  unique?: boolean;
  primary?: boolean;
  foreign?: boolean;
}

export interface FieldDisplay {
  id: string;
  name: string;
  icon: string;
  description: string;
  component: string;
  types: string[];
  options?: any;
}

// Field interfaces available in Directus
export const FIELD_INTERFACES: FieldInterface[] = [
  // Text inputs
  {
    id: 'input',
    name: 'Text Input',
    icon: 'text_fields',
    description: 'Simple text input field',
    component: 'interface-input',
    types: ['string', 'varchar', 'text'],
    groups: ['standard']
  },
  {
    id: 'input-rich-text-html',
    name: 'WYSIWYG',
    icon: 'format_bold',
    description: 'Rich text editor with HTML output',
    component: 'interface-input-rich-text-html',
    types: ['text'],
    groups: ['standard']
  },
  {
    id: 'input-rich-text-md',
    name: 'Markdown',
    icon: 'integration_instructions',
    description: 'Markdown editor',
    component: 'interface-input-rich-text-md',
    types: ['text'],
    groups: ['standard']
  },
  {
    id: 'textarea',
    name: 'Textarea',
    icon: 'text_snippet',
    description: 'Multi-line text input',
    component: 'interface-textarea',
    types: ['text'],
    groups: ['standard']
  },
  {
    id: 'input-code',
    name: 'Code',
    icon: 'code',
    description: 'Code editor with syntax highlighting',
    component: 'interface-input-code',
    types: ['text', 'json'],
    groups: ['standard']
  },

  // Numbers
  {
    id: 'input-number',
    name: 'Number',
    icon: 'pin',
    description: 'Numeric input field',
    component: 'interface-input-number',
    types: ['integer', 'bigint', 'float', 'decimal'],
    groups: ['standard']
  },
  {
    id: 'slider',
    name: 'Slider',
    icon: 'linear_scale',
    description: 'Numeric slider input',
    component: 'interface-slider',
    types: ['integer', 'float'],
    groups: ['standard']
  },

  // Boolean
  {
    id: 'boolean',
    name: 'Boolean',
    icon: 'check_box',
    description: 'Toggle switch for true/false values',
    component: 'interface-boolean',
    types: ['boolean'],
    groups: ['standard']
  },

  // Date and Time
  {
    id: 'datetime',
    name: 'Date & Time',
    icon: 'today',
    description: 'Date and time picker',
    component: 'interface-datetime',
    types: ['timestamp', 'datetime'],
    groups: ['standard']
  },
  {
    id: 'date',
    name: 'Date',
    icon: 'date_range',
    description: 'Date picker',
    component: 'interface-date',
    types: ['date'],
    groups: ['standard']
  },
  {
    id: 'time',
    name: 'Time',
    icon: 'schedule',
    description: 'Time picker',
    component: 'interface-time',
    types: ['time'],
    groups: ['standard']
  },

  // Selection
  {
    id: 'select-dropdown',
    name: 'Dropdown',
    icon: 'arrow_drop_down_circle',
    description: 'Dropdown selection',
    component: 'interface-select-dropdown',
    types: ['string', 'integer'],
    groups: ['selection']
  },
  {
    id: 'select-dropdown-m2o',
    name: 'Dropdown (Many-to-One)',
    icon: 'call_merge',
    description: 'Many-to-one relationship dropdown',
    component: 'interface-select-dropdown-m2o',
    types: ['uuid', 'string', 'integer'],
    groups: ['relational']
  },
  {
    id: 'select-radio',
    name: 'Radio Buttons',
    icon: 'radio_button_checked',
    description: 'Radio button selection',
    component: 'interface-select-radio',
    types: ['string', 'integer'],
    groups: ['selection']
  },
  {
    id: 'checkboxes',
    name: 'Checkboxes',
    icon: 'check_box',
    description: 'Multiple checkbox selection',
    component: 'interface-checkboxes',
    types: ['json'],
    groups: ['selection']
  },

  // Files and Media
  {
    id: 'file-image',
    name: 'Image',
    icon: 'image',
    description: 'Image file upload',
    component: 'interface-file-image',
    types: ['uuid'],
    groups: ['files']
  },
  {
    id: 'file',
    name: 'File',
    icon: 'attach_file',
    description: 'File upload',
    component: 'interface-file',
    types: ['uuid'],
    groups: ['files']
  },
  {
    id: 'files',
    name: 'Files',
    icon: 'folder_open',
    description: 'Multiple file upload',
    component: 'interface-files',
    types: ['json'],
    groups: ['files']
  },

  // JSON and Arrays
  {
    id: 'input-hash',
    name: 'Key-Value Pairs',
    icon: 'code',
    description: 'Key-value pair editor',
    component: 'interface-input-hash',
    types: ['json'],
    groups: ['other']
  },
  {
    id: 'tags',
    name: 'Tags',
    icon: 'local_offer',
    description: 'Tag input field',
    component: 'interface-tags',
    types: ['json'],
    groups: ['other']
  },

  // Presentation
  {
    id: 'presentation-divider',
    name: 'Divider',
    icon: 'remove',
    description: 'Visual divider (no data)',
    component: 'interface-presentation-divider',
    types: [],
    groups: ['presentation'],
    system: true
  },
  {
    id: 'presentation-notice',
    name: 'Notice',
    icon: 'info',
    description: 'Information notice (no data)',
    component: 'interface-presentation-notice',
    types: [],
    groups: ['presentation'],
    system: true
  },

  // System
  {
    id: 'input-uuid',
    name: 'UUID',
    icon: 'fingerprint',
    description: 'UUID input field',
    component: 'interface-input-uuid',
    types: ['uuid'],
    groups: ['other']
  },
  {
    id: 'input-color',
    name: 'Color',
    icon: 'palette',
    description: 'Color picker',
    component: 'interface-input-color',
    types: ['string'],
    groups: ['other']
  }
];

// Field data types
export const FIELD_TYPES: FieldType[] = [
  {
    id: 'string',
    name: 'String',
    icon: 'text_fields',
    description: 'Variable length text',
    sql: 'VARCHAR',
    length: true,
    nullable: true,
    default: true,
    unique: true
  },
  {
    id: 'text',
    name: 'Text',
    icon: 'text_snippet',
    description: 'Long form text',
    sql: 'TEXT',
    nullable: true,
    default: true
  },
  {
    id: 'integer',
    name: 'Integer',
    icon: 'pin',
    description: 'Whole numbers',
    sql: 'INTEGER',
    nullable: true,
    default: true,
    unique: true,
    primary: true
  },
  {
    id: 'bigint',
    name: 'Big Integer',
    icon: 'pin',
    description: 'Large whole numbers',
    sql: 'BIGINT',
    nullable: true,
    default: true,
    unique: true,
    primary: true
  },
  {
    id: 'float',
    name: 'Float',
    icon: 'pin',
    description: 'Decimal numbers',
    sql: 'DECIMAL',
    decimal: true,
    nullable: true,
    default: true
  },
  {
    id: 'boolean',
    name: 'Boolean',
    icon: 'check_box',
    description: 'True or false',
    sql: 'BOOLEAN',
    nullable: true,
    default: true
  },
  {
    id: 'date',
    name: 'Date',
    icon: 'date_range',
    description: 'Date only',
    sql: 'DATE',
    nullable: true,
    default: true
  },
  {
    id: 'time',
    name: 'Time',
    icon: 'schedule',
    description: 'Time only',
    sql: 'TIME',
    nullable: true,
    default: true
  },
  {
    id: 'timestamp',
    name: 'Timestamp',
    icon: 'today',
    description: 'Date and time',
    sql: 'TIMESTAMP',
    nullable: true,
    default: true
  },
  {
    id: 'uuid',
    name: 'UUID',
    icon: 'fingerprint',
    description: 'Universally unique identifier',
    sql: 'UUID',
    nullable: true,
    default: true,
    unique: true,
    primary: true,
    foreign: true
  },
  {
    id: 'json',
    name: 'JSON',
    icon: 'code',
    description: 'JSON data structure',
    sql: 'JSONB',
    nullable: true,
    default: true
  }
];

// Field display options
export const FIELD_DISPLAYS: FieldDisplay[] = [
  {
    id: 'raw',
    name: 'Raw Value',
    icon: 'text_fields',
    description: 'Display the raw value',
    component: 'display-raw',
    types: ['string', 'text', 'integer', 'float']
  },
  {
    id: 'formatted-value',
    name: 'Formatted Value',
    icon: 'text_format',
    description: 'Display formatted value',
    component: 'display-formatted-value',
    types: ['integer', 'float', 'date', 'timestamp']
  },
  {
    id: 'boolean',
    name: 'Boolean',
    icon: 'check_circle',
    description: 'Display as check/cross icon',
    component: 'display-boolean',
    types: ['boolean']
  },
  {
    id: 'datetime',
    name: 'Date & Time',
    icon: 'today',
    description: 'Display formatted date and time',
    component: 'display-datetime',
    types: ['date', 'time', 'timestamp']
  },
  {
    id: 'user',
    name: 'User',
    icon: 'person',
    description: 'Display user information',
    component: 'display-user',
    types: ['uuid']
  },
  {
    id: 'image',
    name: 'Image',
    icon: 'image',
    description: 'Display image thumbnail',
    component: 'display-image',
    types: ['uuid']
  },
  {
    id: 'file',
    name: 'File',
    icon: 'attach_file',
    description: 'Display file information',
    component: 'display-file',
    types: ['uuid']
  },
  {
    id: 'json',
    name: 'JSON',
    icon: 'code',
    description: 'Display formatted JSON',
    component: 'display-json',
    types: ['json']
  },
  {
    id: 'tags',
    name: 'Tags',
    icon: 'local_offer',
    description: 'Display as tags',
    component: 'display-tags',
    types: ['json']
  },
  {
    id: 'color',
    name: 'Color',
    icon: 'palette',
    description: 'Display color swatch',
    component: 'display-color',
    types: ['string']
  }
];

// Field validation rules
export const VALIDATION_RULES = {
  required: {
    name: 'Required',
    description: 'Field must have a value',
    types: ['string', 'text', 'integer', 'float', 'boolean', 'date', 'timestamp', 'uuid', 'json']
  },
  unique: {
    name: 'Unique',
    description: 'Value must be unique across all items',
    types: ['string', 'text', 'integer', 'float', 'uuid']
  },
  length: {
    name: 'Length',
    description: 'Text length constraints',
    types: ['string', 'text'],
    options: ['min', 'max']
  },
  range: {
    name: 'Range',
    description: 'Numeric range constraints',
    types: ['integer', 'float'],
    options: ['min', 'max']
  },
  regex: {
    name: 'Regular Expression',
    description: 'Pattern matching validation',
    types: ['string', 'text'],
    options: ['pattern', 'flags']
  },
  email: {
    name: 'Email',
    description: 'Valid email address format',
    types: ['string']
  },
  url: {
    name: 'URL',
    description: 'Valid URL format',
    types: ['string']
  },
  date_range: {
    name: 'Date Range',
    description: 'Date range constraints',
    types: ['date', 'timestamp'],
    options: ['min', 'max']
  }
};

// Utility functions
export class FieldsService {
  // Get available interfaces for a field type
  static getInterfacesForType(fieldType: string): FieldInterface[] {
    return FIELD_INTERFACES.filter(interface_ => 
      interface_.types.includes(fieldType) || interface_.types.length === 0
    );
  }

  // Get available displays for a field type
  static getDisplaysForType(fieldType: string): FieldDisplay[] {
    return FIELD_DISPLAYS.filter(display => 
      display.types.includes(fieldType)
    );
  }

  // Get field type by ID
  static getFieldType(typeId: string): FieldType | undefined {
    return FIELD_TYPES.find(type => type.id === typeId);
  }

  // Get field interface by ID
  static getFieldInterface(interfaceId: string): FieldInterface | undefined {
    return FIELD_INTERFACES.find(interface_ => interface_.id === interfaceId);
  }

  // Get field display by ID
  static getFieldDisplay(displayId: string): FieldDisplay | undefined {
    return FIELD_DISPLAYS.find(display => display.id === displayId);
  }

  // Validate field configuration
  static validateField(field: CreateFieldRequest | UpdateFieldRequest): string[] {
    const errors: string[] = [];

    // Validate field name (only for create)
    if ('field' in field) {
      if (!field.field || field.field.trim() === '') {
        errors.push('Field name is required');
      } else if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(field.field)) {
        errors.push('Field name must start with a letter and contain only letters, numbers, and underscores');
      }
    }

    // Validate interface and type compatibility
    if (field.interface && field.schema?.data_type) {
      const interface_ = this.getFieldInterface(field.interface);
      if (interface_ && !interface_.types.includes(field.schema.data_type)) {
        errors.push(`Interface "${field.interface}" is not compatible with type "${field.schema.data_type}"`);
      }
    }

    // Validate display and type compatibility
    if (field.display && field.schema?.data_type) {
      const display = this.getFieldDisplay(field.display);
      if (display && !display.types.includes(field.schema.data_type)) {
        errors.push(`Display "${field.display}" is not compatible with type "${field.schema.data_type}"`);
      }
    }

    // Validate schema constraints
    if (field.schema) {
      if (field.schema.max_length && field.schema.max_length <= 0) {
        errors.push('Maximum length must be greater than 0');
      }
      if (field.schema.foreign_table && !field.schema.foreign_column) {
        errors.push('Foreign column is required when foreign table is specified');
      }
      if (field.schema.foreign_column && !field.schema.foreign_table) {
        errors.push('Foreign table is required when foreign column is specified');
      }
    }

    return errors;
  }

  // Generate default field configuration for a type
  static getDefaultFieldConfig(fieldType: string, fieldName: string): Partial<CreateFieldRequest> {
    const type = this.getFieldType(fieldType);
    if (!type) {
      throw new Error(`Unknown field type: ${fieldType}`);
    }

    const availableInterfaces = this.getInterfacesForType(fieldType);
    const defaultInterface = availableInterfaces[0];

    const availableDisplays = this.getDisplaysForType(fieldType);
    const defaultDisplay = availableDisplays[0];

    const config: Partial<CreateFieldRequest> = {
      field: fieldName,
      interface: defaultInterface?.id,
      display: defaultDisplay?.id,
      width: 'full',
      readonly: false,
      hidden: false,
      required: false,
      schema: {
        data_type: fieldType,
        is_nullable: true
      }
    };

    // Type-specific defaults
    switch (fieldType) {
      case 'string':
        config.schema!.max_length = 255;
        break;
      case 'boolean':
        config.schema!.default_value = false;
        break;
      case 'uuid':
        if (fieldName === 'id') {
          config.schema!.is_primary_key = true;
          config.schema!.is_nullable = false;
          config.schema!.default_value = 'gen_random_uuid()';
        }
        break;
      case 'timestamp':
        if (fieldName.includes('created') || fieldName.includes('updated')) {
          config.schema!.default_value = 'CURRENT_TIMESTAMP';
          config.schema!.is_nullable = false;
        }
        break;
    }

    return config;
  }

  // Check if field is a system field
  static isSystemField(fieldName: string): boolean {
    const systemFields = ['id', 'created_at', 'updated_at'];
    return systemFields.includes(fieldName);
  }

  // Check if interface is virtual (doesn't need database column)
  static isVirtualInterface(interfaceId?: string): boolean {
    if (!interfaceId) return false;
    const virtualInterfaces = ['presentation-divider', 'presentation-notice', 'group-raw', 'group-detail', 'alias'];
    return virtualInterfaces.includes(interfaceId);
  }

  // Wrapper methods that delegate to collectionsService
  static async getFields(page: number = 1, limit: number = 50, collection?: string) {
    return collectionsService.getFields(page, limit, collection);
  }

  static async getFieldsByCollection(collectionName: string) {
    return collectionsService.getFieldsByCollection(collectionName);
  }

  static async getField(collectionName: string, fieldName: string) {
    return collectionsService.getField(collectionName, fieldName);
  }

  static async createField(collectionName: string, fieldData: CreateFieldRequest) {
    return collectionsService.createField(collectionName, fieldData);
  }

  static async updateField(collectionName: string, fieldName: string, fieldData: UpdateFieldRequest) {
    return collectionsService.updateField(collectionName, fieldName, fieldData);
  }

  static async deleteField(collectionName: string, fieldName: string) {
    return collectionsService.deleteField(collectionName, fieldName);
  }
}

export default FieldsService;
