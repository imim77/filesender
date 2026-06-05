import { useCallback, useEffect, useRef, useState } from 'react';

type UseSidebarResizeOptions = {
  defaultWidth?: number;
  minWidth?: number;
  maxWidth?: number;
  collapseThreshold?: number;
  expandTrigger?: number;
  overlayStartWidth?: number;
};

export function useSidebarResize({
  defaultWidth = 300,
  minWidth = 200,
  maxWidth = 480,
  collapseThreshold = 180,
  expandTrigger = 50,
  overlayStartWidth = 260,
}: UseSidebarResizeOptions = {}) {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isDragging, setIsDragging] = useState(false);

  const widthRef = useRef(defaultWidth);
  const collapsedRef = useRef(false);
  const draggingRef = useRef(false);
  const startXRef = useRef(0);
  const startWidthRef = useRef(0);
  const rafRef = useRef<number>(0);

  const sidebarRef = useRef<HTMLDivElement>(null);
  const overlayRef = useRef<HTMLDivElement>(null);

  const applyWidth = (w: number) => {
    if (sidebarRef.current) {
      sidebarRef.current.style.width = `${w}px`;
    }
    if (overlayRef.current) {
      if (w >= overlayStartWidth) {
        overlayRef.current.style.opacity = '0';
      } else if (w <= collapseThreshold) {
        overlayRef.current.style.opacity = '0.75';
      } else {
        const progress = 1 - (w - collapseThreshold) / (overlayStartWidth - collapseThreshold);
        overlayRef.current.style.opacity = String(progress * 0.75);
      }
    }
  };

  const setTransition = (on: boolean) => {
    const t = on ? 'width 500ms cubic-bezier(0.2, 0, 0, 1)' : 'none';
    const tO = on ? 'opacity 500ms cubic-bezier(0.2, 0, 0, 1)' : 'none';
    if (sidebarRef.current) sidebarRef.current.style.transition = t;
    if (overlayRef.current) overlayRef.current.style.transition = tO;
  };

  const collapse = () => {
    collapsedRef.current = true;
    setIsCollapsed(true);
    setTransition(true);
    if (sidebarRef.current) sidebarRef.current.style.width = '0px';
    if (overlayRef.current) overlayRef.current.style.opacity = '0.75';
  };

  const expand = () => {
    collapsedRef.current = false;
    widthRef.current = defaultWidth;
    setIsCollapsed(false);
    setTransition(true);
    if (sidebarRef.current) sidebarRef.current.style.width = `${defaultWidth}px`;
    if (overlayRef.current) overlayRef.current.style.opacity = '0';
  };

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    draggingRef.current = true;
    startXRef.current = e.clientX;
    startWidthRef.current = collapsedRef.current ? 0 : widthRef.current;
    setTransition(false);
    setIsDragging(true);
  }, []);

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!draggingRef.current) return;
      if (rafRef.current) cancelAnimationFrame(rafRef.current);

      rafRef.current = requestAnimationFrame(() => {
        const delta = e.clientX - startXRef.current;
        const newWidth = startWidthRef.current + delta;

        if (collapsedRef.current) {
          // When collapsed, just need a small drag right to trigger expand
          if (delta > expandTrigger) {
            expand();
            draggingRef.current = false;
            setIsDragging(false);
          }
        } else {
          if (newWidth < collapseThreshold) {
            collapse();
            draggingRef.current = false;
            setIsDragging(false);
          } else {
            const clamped = Math.min(Math.max(newWidth, minWidth), maxWidth);
            widthRef.current = clamped;
            applyWidth(clamped);
          }
        }
      });
    };

    const handleMouseUp = () => {
      if (!draggingRef.current) return;
      draggingRef.current = false;
      setIsDragging(false);
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [minWidth, maxWidth, collapseThreshold, expandTrigger, overlayStartWidth, defaultWidth]);

  const handleDoubleClick = useCallback(() => {
    if (collapsedRef.current) expand();
    else collapse();
  }, [defaultWidth]);

  return { isCollapsed, isDragging, handleMouseDown, handleDoubleClick, sidebarRef, overlayRef };
}
