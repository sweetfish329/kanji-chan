<script lang="ts">
  import { DatePicker } from 'bits-ui';
  import { parseDate, CalendarDate, type DateValue } from '@internationalized/date';

  interface Props {
    value?: string; // "YYYY-MM-DD"
    placeholder?: string;
    disabled?: boolean;
    class?: string;
    onValueChange?: (val: string) => void;
  }

  let {
    value = $bindable(''),
    placeholder = '日付を選択...',
    disabled = false,
    class: className = '',
    onValueChange
  }: Props = $props();

  // Bits UI DatePicker 用の DateValue に変換
  let parsedDateValue = $derived.by<DateValue | undefined>(() => {
    if (!value) return undefined;
    try {
      const [y, m, d] = value.split('-').map(Number);
      if (y && m && d) {
        return new CalendarDate(y, m, d);
      }
    } catch {
      return undefined;
    }
    return undefined;
  });

  function handleDateChange(val: DateValue | undefined) {
    if (val) {
      const year = val.year;
      const month = String(val.month).padStart(2, '0');
      const day = String(val.day).padStart(2, '0');
      const formatted = `${year}-${month}-${day}`;
      value = formatted;
      onValueChange?.(formatted);
    } else {
      value = '';
      onValueChange?.('');
    }
  }
</script>

<div class="bits-datepicker-container {className}">
  <DatePicker.Root value={parsedDateValue} onValueChange={handleDateChange} {disabled}>
    <div class="bits-datepicker-input-wrapper">
      <DatePicker.Input class="bits-datepicker-input" placeholder={placeholder}>
        {#snippet child({ props })}
          <input
            {...props}
            type="date"
            class="bits-native-date-input"
            bind:value
            {disabled}
            onchange={(e) => {
              const val = (e.target as HTMLInputElement).value;
              value = val;
              onValueChange?.(val);
            }}
          />
        {/snippet}
      </DatePicker.Input>

      <DatePicker.Trigger class="bits-datepicker-btn" aria-label="カレンダーを開く" {disabled}>
        <span class="material-symbols-rounded" aria-hidden="true">calendar_today</span>
      </DatePicker.Trigger>
    </div>

    <DatePicker.Portal>
      <DatePicker.Content sideOffset={6} class="bits-datepicker-content">
        <DatePicker.Calendar class="bits-datepicker-calendar">
          {#snippet child({ months, weekdays })}
            <div class="bits-datepicker-header">
              <DatePicker.PrevButton class="bits-datepicker-nav-btn">
                <span class="material-symbols-rounded" aria-hidden="true">chevron_left</span>
              </DatePicker.PrevButton>
              <DatePicker.Heading class="bits-datepicker-heading" />
              <DatePicker.NextButton class="bits-datepicker-nav-btn">
                <span class="material-symbols-rounded" aria-hidden="true">chevron_right</span>
              </DatePicker.NextButton>
            </div>

            {#each months as month}
              <DatePicker.Grid class="bits-datepicker-grid">
                <DatePicker.GridHead>
                  <DatePicker.GridRow class="bits-datepicker-weekdays">
                    {#each weekdays as day}
                      <DatePicker.HeadCell class="bits-datepicker-weekday">
                        {day}
                      </DatePicker.HeadCell>
                    {/each}
                  </DatePicker.GridRow>
                </DatePicker.GridHead>
                <DatePicker.GridBody>
                  {#each month.weeks as weekDates}
                    <DatePicker.GridRow class="bits-datepicker-row">
                      {#each weekDates as date}
                        <DatePicker.Cell {date} month={month.value} class="bits-datepicker-cell">
                          <DatePicker.Day class="bits-datepicker-day">
                            {date.day}
                          </DatePicker.Day>
                        </DatePicker.Cell>
                      {/each}
                    </DatePicker.GridRow>
                  {/each}
                </DatePicker.GridBody>
              </DatePicker.Grid>
            {/each}
          {/snippet}
        </DatePicker.Calendar>
      </DatePicker.Content>
    </DatePicker.Portal>
  </DatePicker.Root>
</div>

<style>
  .bits-datepicker-container {
    display: inline-flex;
    position: relative;
    width: 100%;
  }

  .bits-datepicker-input-wrapper {
    display: flex;
    align-items: center;
    width: 100%;
    position: relative;
  }

  .bits-native-date-input {
    width: 100%;
    padding: 0.65rem 2.5rem 0.65rem 0.85rem;
    font-family: inherit;
    font-size: 0.9rem;
    color: var(--text-primary);
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    outline: none;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
  }

  .bits-native-date-input::-webkit-calendar-picker-indicator {
    display: none;
    -webkit-appearance: none;
  }

  .bits-native-date-input:focus {
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px rgba(45, 179, 103, 0.18);
  }

  :global(.bits-datepicker-btn) {
    position: absolute;
    right: 0.5rem;
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 0.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-sm);
    transition: color 0.15s ease;
  }

  :global(.bits-datepicker-btn:hover) {
    color: var(--color-primary);
  }

  :global(.bits-datepicker-content) {
    z-index: 550;
    background: var(--bg-glass);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    padding: 1rem;
    animation: datepickerFade 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes datepickerFade {
    from {
      opacity: 0;
      transform: scale(0.96);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  .bits-datepicker-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.75rem;
  }

  :global(.bits-datepicker-heading) {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  :global(.bits-datepicker-nav-btn) {
    background: none;
    border: 1px solid var(--border-glass);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    padding: 0.25rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease;
  }

  :global(.bits-datepicker-nav-btn:hover) {
    background: rgba(45, 179, 103, 0.1);
    color: var(--color-primary);
  }

  .bits-datepicker-grid {
    width: 100%;
    border-collapse: collapse;
  }

  .bits-datepicker-weekdays {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    margin-bottom: 0.5rem;
  }

  :global(.bits-datepicker-weekday) {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text-muted);
    text-align: center;
  }

  .bits-datepicker-row {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 0.2rem;
  }

  :global(.bits-datepicker-day) {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    font-size: 0.85rem;
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease;
  }

  :global(.bits-datepicker-day:hover) {
    background: rgba(45, 179, 103, 0.15);
  }

  :global(.bits-datepicker-day[data-selected]) {
    background: var(--color-primary);
    color: #ffffff;
    font-weight: 600;
  }
</style>
