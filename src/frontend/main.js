import { addUploadButtonListeners } from './FileUpload.js';
import { addDeleteButtonListeners } from './delete.js';
import { addExportButtonListeners, addSendButtonListener, addTagListeners } from './notes.js';

addUploadButtonListeners();
addDeleteButtonListeners();
addExportButtonListeners();
addSendButtonListener();
addTagListeners();
