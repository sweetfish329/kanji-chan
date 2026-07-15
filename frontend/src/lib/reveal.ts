export function reveal(node: HTMLElement, options = { threshold: 0.1, rootMargin: '0px 0px -50px 0px' }) {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        node.classList.add('visible');
        observer.unobserve(node);
      }
    });
  }, options);

  // 初期クラス追加
  node.classList.add('reveal-on-scroll');
  
  // わずかに遅延させて初期状態のレンダリング競合を回避
  setTimeout(() => {
    observer.observe(node);
  }, 50);

  return {
    destroy() {
      observer.disconnect();
    }
  };
}
