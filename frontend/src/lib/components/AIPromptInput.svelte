<script lang="ts">
  import Tooltip from './ui/Tooltip.svelte';

  export interface AttachedImage {
    id: string;
    data: string;      // Base64 Data URL (e.g. data:image/png;base64,...)
    mime_type: string; // e.g. image/png
    name: string;
    sizeFormatted?: string;
  }

  interface TemplateItem {
    label: string;
    text: string;
  }

  interface Props {
    prompt?: string;
    images?: AttachedImage[];
    placeholder?: string;
    submitLabel?: string;
    isSubmitting?: boolean;
    disabled?: boolean;
    templates?: TemplateItem[];
    onSubmit: () => void;
  }

  let {
    prompt = $bindable(''),
    images = $bindable([] as AttachedImage[]),
    placeholder = 'AIへ伝える指示や希望条件を入力...',
    submitLabel = 'AIで分析・生成する',
    isSubmitting = false,
    disabled = false,
    templates = [],
    onSubmit
  }: Props = $props();

  let fileInputRef: HTMLInputElement | undefined = $state();
  let isDragging = $state(false);

  // ファイル選択
  function handleFileSelect(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      processFiles(Array.from(target.files));
      target.value = '';
    }
  }

  // ドラッグ＆ドロップ
  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    isDragging = true;
  }

  function handleDragLeave(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
    if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
      processFiles(Array.from(e.dataTransfer.files));
    }
  }

  // クリップボードからのペースト (Ctrl+V / Cmd+V)
  function handlePaste(e: ClipboardEvent) {
    if (!e.clipboardData?.items) return;
    const items = Array.from(e.clipboardData.items);
    const imageFiles: File[] = [];

    for (const item of items) {
      if (item.type.startsWith('image/')) {
        const file = item.getAsFile();
        if (file) imageFiles.push(file);
      }
    }

    if (imageFiles.length > 0) {
      processFiles(imageFiles);
    }
  }

  // 画像ファイルの読み込み・Base64化
  function processFiles(files: File[]) {
    files.forEach(file => {
      if (!file.type.startsWith('image/')) return;
      if (file.size > 10 * 1024 * 1024) {
        alert('画像サイズは10MB以下にしてください。');
        return;
      }

      const reader = new FileReader();
      reader.onload = (evt) => {
        const dataUrl = evt.target?.result as string;
        if (dataUrl) {
          const newImg: AttachedImage = {
            id: Math.random().toString(36).substring(2, 11),
            data: dataUrl,
            mime_type: file.type,
            name: file.name,
            sizeFormatted: (file.size / 1024).toFixed(0) + 'KB'
          };
          images = [...images, newImg];
        }
      };
      reader.readAsDataURL(file);
    });
  }

  // 添付画像の削除
  function removeImage(id: string) {
    images = images.filter(img => img.id !== id);
  }

  // テンプレート入力セット
  function applyTemplate(tplText: string) {
    prompt = tplText;
  }

  // キーボード送信 (Enter送信, Shift+Enter改行)
  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      if (!isSubmitting && !disabled && (prompt.trim() || images.length > 0)) {
        onSubmit();
      }
    }
  }
</script>

<div 
    class="ai-prompt-container" 
    class:is-dragging={isDragging}
    role="region"
    aria-label="AIプロンプト入力および画像ドロップエリア"
    ondragover={handleDragOver}
    ondragleave={handleDragLeave}
    ondrop={handleDrop}
  >
    <!-- Dropzone overlay indicator -->
    {#if isDragging}
      <div class="drag-overlay">
        <span class="material-symbols-rounded drop-icon">cloud_upload</span>
        <p>ここに画像をドロップして添付</p>
      </div>
    {/if}

    <!-- Templates chips -->
    {#if templates.length > 0}
      <div class="template-chips-row">
        <span class="chips-label">💡 クイック例:</span>
        <div class="chips-scroll">
          {#each templates as tpl}
            <button 
              type="button" 
              class="chip-btn"
              disabled={disabled || isSubmitting}
              onclick={() => applyTemplate(tpl.text)}
            >
              {tpl.label}
            </button>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Image Attachments Preview Grid -->
    {#if images.length > 0}
      <div class="attachments-preview-grid">
        {#each images as img (img.id)}
          <div class="img-preview-card">
            <img src={img.data} alt={img.name} class="preview-thumbnail" />
            <div class="preview-meta">
              <span class="img-name">{img.name}</span>
              {#if img.sizeFormatted}
                <span class="img-size">{img.sizeFormatted}</span>
              {/if}
            </div>
            
            <Tooltip content="添付画像を削除" side="top">
              {#snippet trigger()}
                <button type="button" class="btn-remove-img" aria-label={`画像 ${img.name} を削除`} onclick={() => removeImage(img.id)}>
                  <span class="material-symbols-rounded">close</span>
                </button>
              {/snippet}
            </Tooltip>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Prompt Textarea & Actions Bar -->
    <div class="prompt-input-wrapper">
      <textarea
        bind:value={prompt}
        {placeholder}
        {disabled}
        rows="3"
        class="prompt-textarea"
        onkeydown={handleKeyDown}
        onpaste={handlePaste}
      ></textarea>

      <!-- Bottom Actions Toolbar -->
      <div class="input-actions-bar">
        <div class="left-tools">
          <!-- Image Attach Button with Bits UI Tooltip -->
          <Tooltip content="チラシ・カレンダー・メモ写真を添付" side="top">
            {#snippet trigger()}
              <button
                type="button" 
                class="tool-btn attachment-btn"
                disabled={disabled || isSubmitting}
                onclick={() => fileInputRef?.click()}
              >
                <span class="material-symbols-rounded">add_photo_alternate</span>
                <span class="tool-label">画像添付</span>
              </button>
            {/snippet}
          </Tooltip>

          <input 
            type="file" 
            accept="image/*" 
            multiple 
            bind:this={fileInputRef} 
            onchange={handleFileSelect} 
            class="hidden-file-input"
          />

          <span class="hint-text">（ドラッグ＆ドロップ / クリップボード貼り付け対応）</span>
        </div>

        <!-- Submit Button -->
        <button 
          type="button" 
          class="btn btn-primary ai-submit-btn" 
          class:is-submitting={isSubmitting}
          disabled={disabled || isSubmitting || (!prompt.trim() && images.length === 0)}
          onclick={onSubmit}
        >
          <span class="material-symbols-rounded ai-sparkle-icon" class:spin={isSubmitting}>
            {isSubmitting ? 'sync' : 'auto_awesome'}
          </span>
          <span>{isSubmitting ? 'AI処理中...' : submitLabel}</span>
        </button>
      </div>
    </div>
  </div>

<style>
  .ai-prompt-container {
    position: relative;
    border-radius: var(--radius-md);
    background: var(--bg-glass);
    border: 1px solid var(--border-glass);
    box-shadow: var(--shadow-sm);
    padding: 1rem;
    transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
  }

  .ai-prompt-container.is-dragging {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 3px rgba(212, 140, 56, 0.2);
  }

  .drag-overlay {
    position: absolute;
    inset: 0;
    background: rgba(248, 246, 240, 0.92);
    backdrop-filter: blur(8px);
    z-index: 10;
    border-radius: var(--radius-md);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: var(--color-accent);
    pointer-events: none;
    animation: fadeIn 0.15s ease;
  }

  .drop-icon {
    font-size: 2.5rem;
    margin-bottom: 0.4rem;
  }

  /* Templates Chips */
  .template-chips-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .chips-label {
    font-size: 0.78rem;
    color: var(--text-muted);
    font-weight: 600;
    white-space: nowrap;
  }

  .chips-scroll {
    display: flex;
    gap: 0.4rem;
    overflow-x: auto;
    padding-bottom: 0.2rem;
    -webkit-overflow-scrolling: touch;
  }

  .chip-btn {
    background: var(--bg-secondary);
    border: 1px solid var(--border-glass);
    color: var(--color-primary);
    padding: 0.35rem 0.75rem;
    border-radius: var(--radius-full);
    font-size: 0.8rem;
    white-space: nowrap;
    cursor: pointer;
    transition: transform 0.2s ease, background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
  }

  .chip-btn:hover:not(:disabled) {
    background: var(--color-primary);
    color: #fff;
    transform: translateY(-1px);
  }

  /* Image Attachments Preview */
  .attachments-preview-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
    margin-bottom: 0.75rem;
  }

  .img-preview-card {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.6);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    padding: 0.4rem 0.6rem;
    max-width: 220px;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.03);
    animation: scaleIn 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes scaleIn {
    from { opacity: 0; transform: scale(0.9); }
    to { opacity: 1; transform: scale(1); }
  }

  .preview-thumbnail {
    width: 36px;
    height: 36px;
    object-fit: cover;
    border-radius: 6px;
    border: 1px solid rgba(0, 0, 0, 0.08);
  }

  .preview-meta {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    flex: 1;
  }

  .img-name {
    font-size: 0.78rem;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .img-size {
    font-size: 0.7rem;
    color: var(--text-muted);
  }

  :global(.btn-remove-img) {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 0.1rem;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: color 0.15s ease, background 0.15s ease;
  }

  :global(.btn-remove-img:hover) {
    color: var(--color-ng);
    background: rgba(184, 74, 65, 0.1);
  }

  .btn-remove-img .material-symbols-rounded {
    font-size: 1.1rem;
  }

  /* Prompt Input Wrapper */
  .prompt-input-wrapper {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .prompt-textarea {
    width: 100%;
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    padding: 0.8rem 1rem;
    background: #FAF8F5;
    font-family: var(--font-sans);
    font-size: 0.95rem;
    color: var(--text-primary);
    resize: vertical;
    min-height: 80px;
    transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
  }

  .prompt-textarea:focus {
    outline: none;
    border-color: var(--color-accent);
    box-shadow: 0 0 0 3px rgba(212, 140, 56, 0.15);
  }

  /* Toolbar */
  .input-actions-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.6rem;
  }

  .left-tools {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  :global(.tool-btn) {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    background: rgba(42, 64, 50, 0.05);
    border: 1px solid var(--border-glass);
    color: var(--color-primary);
    padding: 0.4rem 0.75rem;
    border-radius: var(--radius-sm);
    font-size: 0.83rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
  }

  :global(.tool-btn:hover:not(:disabled)) {
    background: rgba(42, 64, 50, 0.1);
    color: var(--color-accent);
  }

  .tool-btn .material-symbols-rounded {
    font-size: 1.2rem;
    color: var(--color-accent);
  }

  .hidden-file-input {
    display: none;
  }

  .hint-text {
    font-size: 0.74rem;
    color: var(--text-muted);
  }

  @media (max-width: 640px) {
    .prompt-textarea {
      min-height: 60px;
      padding: 0.6rem 0.8rem;
      font-size: 0.88rem;
    }

    .input-actions-bar {
      flex-direction: column;
      align-items: stretch;
      gap: 0.5rem;
    }

    .left-tools {
      width: 100%;
      justify-content: flex-start;
    }

    .ai-submit-btn {
      width: 100%;
      justify-content: center;
      padding: 0.6rem 1rem;
      font-size: 0.88rem;
    }

    .hint-text {
      display: none;
    }
  }

  .ai-submit-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.6rem 1.25rem;
    font-size: 0.9rem;
  }

  .ai-submit-btn.is-submitting {
    opacity: 0.8;
  }

  .ai-sparkle-icon.spin {
    animation: spin 1.2s infinite linear;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Bits UI Tooltip Styling */
  :global(.bits-tooltip-content) {
    background: rgba(28, 36, 30, 0.92);
    color: #F8F6F0;
    padding: 0.3rem 0.6rem;
    border-radius: 6px;
    font-size: 0.75rem;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    z-index: 1000;
    animation: fadeIn 0.15s ease;
  }
</style>
