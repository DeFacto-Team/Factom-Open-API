import React, { useState } from 'react';
import { Icon, Input, Button } from 'antd';

const InputGroup = Input.Group;

const EditableText = props => {
  const [isEditing, setIsEditing] = useState(false);
  const [text, setText] = useState(props.text);
  const [newText, setNewText] = useState(props.text);
  const [saveDisabled, setSaveDisabled] = useState(false);
  const placeholder = props.placeholder;
  const type = props.type;

  const save = value => {
    if (type === 'number') {
      if (Number.isNaN(Number(value))) {
        value = 0;
      } else {
        value = Number(value);
      }
    }
    props.onSave(value);
    setText(value);
    setIsEditing(false);
  };

  const handleChange = event => {
    if (event.target.value === '') {
      setSaveDisabled(true);
    } else {
      setSaveDisabled(false);
    }
    setNewText(event.target.value);
  };

  if (!isEditing) {
    return (
      <span>
        {text}{' '}
        <a href="javascript:;" onClick={() => setIsEditing(true)}>
          <Icon type="edit" theme="twoTone" />
        </a>
      </span>
    );
  } else {
    return (
      <span>
        <InputGroup compact>
          <Input
            style={{ width: '100px' }}
            placeholder={placeholder}
            defaultValue={text}
            onChange={handleChange}
            onPressEnter={() => save(newText)}
          />
          <Button
            type="danger"
            icon="close"
            onClick={() => setIsEditing(false)}
          />
          <Button
            type="primary"
            icon="check"
            onClick={() => save(newText)}
            disabled={saveDisabled}
          />
        </InputGroup>
      </span>
    );
  }
};

export default EditableText;
