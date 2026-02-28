<script lang="ts">
 import * as Resizable from "$lib/components/ui/resizable/index.js";
	import * as InputGroup from "$lib/components/ui/input-group/index.js";
     import SearchIcon from "@lucide/svelte/icons/search";
</script>
 
<Resizable.PaneGroup direction="horizontal" class="max-w-md rounded-lg border">
 <Resizable.Pane defaultSize={20}>
  <div class="flex h-[500px] items-center justify-center p-6">
   <span class="font-semibold">One</span>
  </div>
 </Resizable.Pane>
 <Resizable.Handle />
 <Resizable.Pane defaultSize={50}>
  <Resizable.PaneGroup direction="vertical">
   <Resizable.Pane defaultSize={25}>
    <div class="flex h-full items-center justify-center p-6">
     <span class="font-semibold">Two</span>
    </div>
      <InputGroup.Root>
    <InputGroup.Input placeholder="Search..." />
    <InputGroup.Addon>
      <SearchIcon />
    </InputGroup.Addon>
    <InputGroup.Addon align="inline-end">12 results</InputGroup.Addon>
  </InputGroup.Root>
   </Resizable.Pane>
   <Resizable.Handle />
   <Resizable.Pane defaultSize={100}>
    <div class="flex h-full items-center justify-center p-6">
     <span class="font-semibold">Three</span>
    </div>
   </Resizable.Pane>
  </Resizable.PaneGroup>
 </Resizable.Pane>
</Resizable.PaneGroup>
