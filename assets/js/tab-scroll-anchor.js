/**
 * Copyright 2026 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Keeps the clicked tab bar at the same screen position when synced tabs
// (params.page.tabs.sync) switch other tab groups and change their height.
(function () {
  const SWITCH_KEYS = ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End'];
  let pending = false; // One measurement per frame; cleared before paint.
  function keepTabInPlace(e) {
    if (e.type === 'keydown' && !SWITCH_KEYS.includes(e.key)) return;
    if (pending || !e.target.closest('.hextra-tabs-toggle')) return;
    const bar = e.currentTarget;
    const initialTop = bar.getBoundingClientRect().top;
    pending = true;
    // Runs after all event handlers and DOM updates, before the next paint.
    requestAnimationFrame(() => {
      pending = false;
      const delta = bar.getBoundingClientRect().top - initialTop;
      if (delta !== 0) window.scrollBy({ top: delta, behavior: 'instant' });
    });
  }
  const bars = new Set(
    Array.from(document.querySelectorAll('.hextra-tabs-toggle'), (t) => t.parentElement)
  );
  bars.forEach((bar) => {
    // Capture runs before hextra's handler on the tab button itself.
    bar.addEventListener('click', keepTabInPlace, true);
    bar.addEventListener('keydown', keepTabInPlace, true);
  });
})();
