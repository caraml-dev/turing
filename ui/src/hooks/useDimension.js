import { useRef, useState, useEffect } from "react";
import ResizeObserver from "resize-observer-polyfill";

const initialState = { width: 0, height: 0 };

const useDimension = ref => {
  const [dimensions, setDimensions] = useState(initialState);
  const resizeObserverRef = useRef(null);

  useEffect(() => {
    resizeObserverRef.current = new ResizeObserver((entries = []) => {
      entries.forEach(entry => {
        const { width, height } = entry.contentRect;
        setDimensions({ width, height });
      });
    });
    if (ref.current) resizeObserverRef.current.observe(ref.current);
    return () => {
      if (resizeObserverRef.current) resizeObserverRef.current.disconnect();
    };
  }, [ref]);
  return dimensions;
};

export default useDimension;
