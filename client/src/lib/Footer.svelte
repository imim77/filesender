

<script>
    import { getAgentInfo } from "../utilis/uaNames";

    const browser = getAgentInfo(navigator.userAgent);
</script>

<footer class="bg-gray-100 w-full mt-8">
    <div class="grid grid-cols-3 gap-4 items-center p-4">
        <div class="text-left">
            It is a long established fact that a reader will be distracted...
        </div>

        <div class="text-center">
            {browser}
        </div>

        <div class="text-right">
            <button class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
                Button
            </button>
        </div>
    </div>
</footer>